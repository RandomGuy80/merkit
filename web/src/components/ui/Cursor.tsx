'use client'

import { useEffect, useState } from 'react'
import { motion, useMotionValue, useSpring } from 'framer-motion'

export function Cursor() {
  const [visible, setVisible] = useState(false)
  const [clicking, setClicking] = useState(false)
  const [hovering, setHovering] = useState(false)

  const mouseX = useMotionValue(-100)
  const mouseY = useMotionValue(-100)

  // Dot follows instantly
  const dotX = useSpring(mouseX, { stiffness: 1000, damping: 60 })
  const dotY = useSpring(mouseY, { stiffness: 1000, damping: 60 })

  // Ring follows with satisfying lag
  const ringX = useSpring(mouseX, { stiffness: 160, damping: 22 })
  const ringY = useSpring(mouseY, { stiffness: 160, damping: 22 })

  useEffect(() => {
    if (typeof window === 'undefined') return
    if ('ontouchstart' in window || navigator.maxTouchPoints > 0) return

    const onMove = (e: MouseEvent) => {
      mouseX.set(e.clientX)
      mouseY.set(e.clientY)
      if (!visible) setVisible(true)

      const target = e.target as HTMLElement
      setHovering(
        !!target.closest('a, button, [role="button"], input, select, textarea, label')
      )
    }
    const onDown = () => setClicking(true)
    const onUp = () => setClicking(false)

    window.addEventListener('mousemove', onMove)
    window.addEventListener('mousedown', onDown)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mousedown', onDown)
      window.removeEventListener('mouseup', onUp)
    }
  }, [mouseX, mouseY, visible])

  if (!visible) return null

  return (
    <>
      {/* Outer ring — lags behind */}
      <motion.div
        className="fixed top-0 left-0 pointer-events-none z-[9998] mix-blend-difference"
        style={{ x: ringX, y: ringY, translateX: '-50%', translateY: '-50%' }}
        animate={{
          scale: clicking ? 0.65 : hovering ? 1.3 : 1,
          opacity: clicking ? 0.5 : 1,
        }}
        transition={{ type: 'spring', stiffness: 400, damping: 28 }}
      >
        <div className="w-9 h-9 rounded-full border border-white" />
      </motion.div>

      {/* Inner dot — instant */}
      <motion.div
        className="fixed top-0 left-0 pointer-events-none z-[9999] mix-blend-difference"
        style={{ x: dotX, y: dotY, translateX: '-50%', translateY: '-50%' }}
        animate={{ scale: clicking ? 0.6 : hovering ? 0 : 1 }}
        transition={{ type: 'spring', stiffness: 600, damping: 35 }}
      >
        <div className="w-[5px] h-[5px] rounded-full bg-white" />
      </motion.div>
    </>
  )
}
