import { useState, useCallback, useRef } from 'react'

interface UseResizableOptions {
  direction: 'left' | 'right'
  minWidth: number
  maxWidth: number
  defaultWidth: number
  storageKey?: string
}

interface UseResizableReturn {
  width: number
  setWidth: (w: number) => void
  onMouseDown: (e: React.MouseEvent) => void
  isResizing: boolean
}

function readStoredWidth(key: string, min: number, max: number): number | null {
  try {
    const raw = localStorage.getItem(key)
    if (raw === null) return null
    const value = Number(raw)
    if (Number.isNaN(value) || value < min || value > max) return null
    return value
  } catch {
    return null
  }
}

export function useResizable({
  direction,
  minWidth,
  maxWidth,
  defaultWidth,
  storageKey,
}: UseResizableOptions): UseResizableReturn {
  const [width, setWidth] = useState(() => {
    if (storageKey) {
      return readStoredWidth(storageKey, minWidth, maxWidth) ?? defaultWidth
    }
    return defaultWidth
  })
  const [isResizing, setIsResizing] = useState(false)
  const startX = useRef(0)
  const startWidth = useRef(0)
  const latestWidth = useRef(0)

  const onMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault()
      startX.current = e.clientX
      startWidth.current = width
      setIsResizing(true)

      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'

      const onMouseMove = (ev: MouseEvent) => {
        const delta = direction === 'left'
          ? startX.current - ev.clientX
          : ev.clientX - startX.current
        const newWidth = Math.min(maxWidth, Math.max(minWidth, startWidth.current + delta))
        latestWidth.current = newWidth
        setWidth(newWidth)
      }

      const onMouseUp = () => {
        document.removeEventListener('mousemove', onMouseMove)
        document.removeEventListener('mouseup', onMouseUp)
        document.body.style.cursor = ''
        document.body.style.userSelect = ''
        setIsResizing(false)

        if (storageKey) {
          try {
            localStorage.setItem(storageKey, String(latestWidth.current))
          } catch {
            // ignore storage errors
          }
        }
      }

      document.addEventListener('mousemove', onMouseMove)
      document.addEventListener('mouseup', onMouseUp)
    },
    [direction, minWidth, maxWidth, width, storageKey],
  )

  return { width, setWidth, onMouseDown, isResizing }
}
