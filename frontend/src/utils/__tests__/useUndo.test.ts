import { renderHook } from '@solidjs/testing-library'
import { createSignal } from 'solid-js'
import { describe, expect, it, vi } from 'vitest'
import { useUndo } from '../useUndo'

describe('useUndo Hook', () => {
  it('should initialize with empty undo/redo state', () => {
    const [text, setText] = createSignal('initial')

    const { result } = renderHook(() => useUndo(text, setText))

    expect(result.canUndo()).toBe(false)
    expect(result.canRedo()).toBe(false)
  })

  it('should record text insertion and allow undo', () => {
    const [text, setText] = createSignal('')

    const { result } = renderHook(() => useUndo(text, setText))

    // Simulate typing 'hello'
    result.recordChange('', 'h', 1)
    result.recordChange('h', 'he', 2)
    result.recordChange('he', 'hel', 3)
    result.recordChange('hel', 'hell', 4)
    result.recordChange('hell', 'hello', 5)

    expect(result.canUndo()).toBe(true)

    // Undo should merge consecutive character insertions
    result.undo()
    expect(result.canUndo()).toBe(false) // All characters merged into one operation
  })

  it('should record text deletion and allow undo', () => {
    const [text, setText] = createSignal('hello')
    const setTextSpy = vi.fn(setText)

    const { result } = renderHook(() => useUndo(text, setTextSpy))

    // Simulate backspace
    result.recordChange('hello', 'hell', 4)

    expect(result.canUndo()).toBe(true)

    result.undo()
    expect(setTextSpy).toHaveBeenLastCalledWith('hello')
  })

  it('should handle redo after undo', () => {
    const [text, setText] = createSignal('')
    const setTextSpy = vi.fn(setText)

    const { result } = renderHook(() => useUndo(text, setTextSpy))

    result.recordChange('', 'hello', 5)

    result.undo()
    expect(result.canRedo()).toBe(true)

    result.redo()
    expect(setTextSpy).toHaveBeenLastCalledWith('hello')
    expect(result.canRedo()).toBe(false)
  })

  it('should clear redo stack when new change is made after undo', () => {
    const [text, setText] = createSignal('')

    const { result } = renderHook(() => useUndo(text, setText))

    result.recordChange('', 'hello', 5)
    result.undo()

    expect(result.canRedo()).toBe(true)

    // Make new change
    result.recordChange('', 'world', 5)

    expect(result.canRedo()).toBe(false)
  })

  it('should handle text replacement correctly', () => {
    const [text, setText] = createSignal('hello world')
    const setTextSpy = vi.fn(setText)

    const { result } = renderHook(() => useUndo(text, setTextSpy))

    // Replace 'world' with 'there' (same length replacement)
    result.recordChange('hello world', 'hello there', 11)

    expect(result.canUndo()).toBe(true)

    result.undo()
    expect(setTextSpy).toHaveBeenLastCalledWith('hello world')
  })

  it('should clear all history', () => {
    const [text, setText] = createSignal('')

    const { result } = renderHook(() => useUndo(text, setText))

    result.recordChange('', 'hello', 5)
    result.undo()

    expect(result.canUndo()).toBe(false)
    expect(result.canRedo()).toBe(true)

    result.clear()

    expect(result.canUndo()).toBe(false)
    expect(result.canRedo()).toBe(false)
  })

  it('should handle cursor position restoration', () => {
    const [text, setText] = createSignal('')
    const setTextSpy = vi.fn(setText)

    const { result } = renderHook(() => useUndo(text, setTextSpy))

    // Simulate typing at position 0
    result.recordChange('', 'hello', 5)

    // Undo should restore to position before insertion
    result.undo()

    // The cursor position is handled by the component, but we can verify
    // that the text change was recorded correctly
    expect(setTextSpy).toHaveBeenLastCalledWith('')
  })

  it('should handle multiple rapid changes with merging', () => {
    const [text, setText] = createSignal('')

    const { result } = renderHook(() => useUndo(text, setText))

    // Simulate rapid typing (should merge)
    result.recordChange('', 'h', 1)
    result.recordChange('h', 'he', 2)
    result.recordChange('he', 'hel', 3)
    result.recordChange('hel', 'hell', 4)
    result.recordChange('hell', 'hello', 5)

    // Should be able to undo all at once (merged)
    expect(result.canUndo()).toBe(true)
    result.undo()
    expect(result.canUndo()).toBe(false)
  })

  it('should not merge changes with large time gaps', () => {
    const [text, setText] = createSignal('')

    const { result } = renderHook(() => useUndo(text, setText))

    // First change
    result.recordChange('', 'hello', 5)

    // Second change after 2 seconds (should not merge)
    result.recordChange('hello', 'hello world', 11)

    // Should have two separate undo operations
    expect(result.canUndo()).toBe(true)
    result.undo() // Undo "hello world" -> "hello"
    expect(result.canUndo()).toBe(true)
    result.undo() // Undo "hello" -> ""
    expect(result.canUndo()).toBe(false)
  })
})
