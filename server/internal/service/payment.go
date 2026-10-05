package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"

	"marketplace/internal/model"
)

type PaymentService struct {
	db            *pgxpool.Pool
	webhookSecret string
}

func NewPaymentService(db *pgxpool.Pool, stripeSecret, webhookSecret string) *PaymentService {
	stripe.Key = stripeSecret
	return &PaymentService{db: db, webhookSecret: webhookSecret}
}

func (s *PaymentService) CreateCheckoutSession(ctx context.Context, orderID, buyerID, successURL, cancelURL string) (string, error) {
	order, err := s.getOrder(ctx, orderID)
	if err != nil {
		return "", err
	}
	if order.BuyerID != buyerID {
		return "", fmt.Errorf("forbidden")
	}
	if order.Status != model.OrderPending {
		return "", fmt.Errorf("order is not pending")
	}

	var listingTitle string
	s.db.QueryRow(ctx, `SELECT title FROM listings WHERE id=$1`, order.ListingID).Scan(&listingTitle)
	if listingTitle == "" {
		listingTitle = "Marketplace Order"
	}

	amountCents := int64(order.Amount * 100)
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(listingTitle),
					},
					UnitAmount: stripe.Int64(amountCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL + "?order_id=" + orderID),
		CancelURL:  stripe.String(cancelURL),
		Metadata:   map[string]string{"order_id": orderID},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`UPDATE orders SET stripe_payment_id=$1 WHERE id=$2`,
		sess.ID, orderID)
	if err != nil {
		return "", fmt.Errorf("save session id: %w", err)
	}

	return sess.URL, nil
}

func (s *PaymentService) HandleWebhook(ctx context.Context, payload []byte, sig string) error {
	event, err := webhook.ConstructEventWithOptions(payload, sig, s.webhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	if event.Type != "checkout.session.completed" {
		return nil
	}

	sess, ok := event.Data.Object["id"].(string)
	if !ok {
		return fmt.Errorf("missing session id in event")
	}

	_, err = s.db.Exec(ctx,
		`UPDATE orders SET status='paid', updated_at=NOW() WHERE stripe_payment_id=$1 AND status='pending'`,
		sess)
	return err
}

func (s *PaymentService) getOrder(ctx context.Context, orderID string) (*model.Order, error) {
	var o model.Order
	err := s.db.QueryRow(ctx,
		`SELECT id, listing_id, buyer_id, seller_id, amount, platform_fee, status, stripe_payment_id, created_at, updated_at
		 FROM orders WHERE id=$1`, orderID,
	).Scan(&o.ID, &o.ListingID, &o.BuyerID, &o.SellerID,
		&o.Amount, &o.PlatformFee, &o.Status, &o.StripePaymentID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}
	return &o, nil
}
