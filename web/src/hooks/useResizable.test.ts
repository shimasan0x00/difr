import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useResizable } from './useResizable'

describe('useResizable', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    localStorage.clear()
  })

  it('returns defaultWidth as initial width', () => {
    const { result } = renderHook(() =>
      useResizable({ direction: 'left', minWidth: 200, maxWidth: 600, defaultWidth: 320 }),
    )

    expect(result.current.width).toBe(320)
    expect(result.current.isResizing).toBe(false)
  })

  it('restores width from localStorage when storageKey is provided', () => {
    localStorage.setItem('difr:chatPanelWidth', '400')

    const { result } = renderHook(() =>
      useResizable({
        direction: 'left',
        minWidth: 200,
        maxWidth: 600,
        defaultWidth: 320,
        storageKey: 'difr:chatPanelWidth',
      }),
    )

    expect(result.current.width).toBe(400)
  })

  it('falls back to defaultWidth when localStorage value is invalid', () => {
    localStorage.setItem('difr:chatPanelWidth', 'not-a-number')

    const { result } = renderHook(() =>
      useResizable({
        direction: 'left',
        minWidth: 200,
        maxWidth: 600,
        defaultWidth: 320,
        storageKey: 'difr:chatPanelWidth',
      }),
    )

    expect(result.current.width).toBe(320)
  })

  it('falls back to defaultWidth when localStorage value is out of range', () => {
    localStorage.setItem('difr:chatPanelWidth', '9999')

    const { result } = renderHook(() =>
      useResizable({
        direction: 'left',
        minWidth: 200,
        maxWidth: 600,
        defaultWidth: 320,
        storageKey: 'difr:chatPanelWidth',
      }),
    )

    expect(result.current.width).toBe(320)
  })

  it('updates width on drag (direction=left: left drag increases width)', () => {
    const { result } = renderHook(() =>
      useResizable({ direction: 'left', minWidth: 200, maxWidth: 600, defaultWidth: 320 }),
    )

    // Start drag at x=500
    act(() => {
      result.current.onMouseDown({ clientX: 500, preventDefault: vi.fn() } as unknown as React.MouseEvent)
    })

    expect(result.current.isResizing).toBe(true)

    // Move mouse left by 80px (x=420) → width increases by 80
    act(() => {
      document.dispatchEvent(new MouseEvent('mousemove', { clientX: 420 }))
    })

    expect(result.current.width).toBe(400)

    // Release mouse
    act(() => {
      document.dispatchEvent(new MouseEvent('mouseup'))
    })

    expect(result.current.isResizing).toBe(false)
  })

  it('clamps width to minWidth', () => {
    const { result } = renderHook(() =>
      useResizable({ direction: 'left', minWidth: 200, maxWidth: 600, defaultWidth: 320 }),
    )

    act(() => {
      result.current.onMouseDown({ clientX: 500, preventDefault: vi.fn() } as unknown as React.MouseEvent)
    })

    // Move right by 200px → width decreases by 200 → 120, but clamped to 200
    act(() => {
      document.dispatchEvent(new MouseEvent('mousemove', { clientX: 700 }))
    })

    expect(result.current.width).toBe(200)
  })

  it('clamps width to maxWidth', () => {
    const { result } = renderHook(() =>
      useResizable({ direction: 'left', minWidth: 200, maxWidth: 600, defaultWidth: 320 }),
    )

    act(() => {
      result.current.onMouseDown({ clientX: 500, preventDefault: vi.fn() } as unknown as React.MouseEvent)
    })

    // Move left by 400px → width increases by 400 → 720, but clamped to 600
    act(() => {
      document.dispatchEvent(new MouseEvent('mousemove', { clientX: 100 }))
    })

    expect(result.current.width).toBe(600)
  })

  it('persists width to localStorage on mouseup when storageKey is provided', () => {
    const { result } = renderHook(() =>
      useResizable({
        direction: 'left',
        minWidth: 200,
        maxWidth: 600,
        defaultWidth: 320,
        storageKey: 'difr:chatPanelWidth',
      }),
    )

    act(() => {
      result.current.onMouseDown({ clientX: 500, preventDefault: vi.fn() } as unknown as React.MouseEvent)
    })

    act(() => {
      document.dispatchEvent(new MouseEvent('mousemove', { clientX: 420 }))
    })

    act(() => {
      document.dispatchEvent(new MouseEvent('mouseup'))
    })

    expect(localStorage.getItem('difr:chatPanelWidth')).toBe('400')
  })

  it('sets cursor and userSelect on body during drag', () => {
    const { result } = renderHook(() =>
      useResizable({ direction: 'left', minWidth: 200, maxWidth: 600, defaultWidth: 320 }),
    )

    act(() => {
      result.current.onMouseDown({ clientX: 500, preventDefault: vi.fn() } as unknown as React.MouseEvent)
    })

    expect(document.body.style.cursor).toBe('col-resize')
    expect(document.body.style.userSelect).toBe('none')

    act(() => {
      document.dispatchEvent(new MouseEvent('mouseup'))
    })

    expect(document.body.style.cursor).toBe('')
    expect(document.body.style.userSelect).toBe('')
  })

  it('works with direction=right (right drag increases width)', () => {
    const { result } = renderHook(() =>
      useResizable({ direction: 'right', minWidth: 200, maxWidth: 600, defaultWidth: 320 }),
    )

    act(() => {
      result.current.onMouseDown({ clientX: 500, preventDefault: vi.fn() } as unknown as React.MouseEvent)
    })

    // Move right by 80px → width increases by 80
    act(() => {
      document.dispatchEvent(new MouseEvent('mousemove', { clientX: 580 }))
    })

    expect(result.current.width).toBe(400)
  })
})
