import { Accessor, createSignal, Setter } from 'solid-js'
import {
  createTextDeleteCommand,
  createTextInsertCommand,
  createUndoStack
} from './undoStack'

export interface UseUndoReturn {
  undo: () => void
  redo: () => void
  canUndo: Accessor<boolean>
  canRedo: Accessor<boolean>
  recordChange: (oldText: string, newText: string, cursorPos: number) => void
  clear: () => void
}

export function useUndo(text: Accessor<string>, setText: Setter<string>): UseUndoReturn {
  const undoStack = createUndoStack()
  const [canUndo, setCanUndo] = createSignal(false)
  const [canRedo, setCanRedo] = createSignal(false)

  let isUndoRedo = false

  const updateCanUndoRedo = () => {
    setCanUndo(undoStack.canUndo())
    setCanRedo(undoStack.canRedo())
  }

  const recordChange = (oldText: string, newText: string, cursorPos: number) => {
    if (isUndoRedo || oldText === newText) return

    const oldLen = oldText.length
    const newLen = newText.length

    const wrappedSetText = (text: string) => {
      isUndoRedo = true
      setText(text)
      isUndoRedo = false
    }

    if (newLen > oldLen) {
      // Text was inserted
      const insertedText = newText.slice(cursorPos - (newLen - oldLen), cursorPos)
      const position = cursorPos - insertedText.length

      const command = createTextInsertCommand(
        wrappedSetText,
        text,
        insertedText,
        position
      )

      undoStack.addToStack(command)
    } else if (newLen < oldLen) {
      // Text was deleted
      const deletedText = oldText.slice(cursorPos, cursorPos + (oldLen - newLen))

      const command = createTextDeleteCommand(
        wrappedSetText,
        () => newText,
        deletedText,
        cursorPos
      )

      undoStack.addToStack(command)
    } else if (newLen === oldLen && oldText !== newText) {
      // Text was replaced with same length (e.g., "world" -> "there")
      // We need to find the range that changed and create a compound command
      let start = 0
      let end = oldLen

      // Find the first differing character
      while (start < oldLen && oldText[start] === newText[start]) {
        start++
      }

      // Find the last differing character
      while (end > start && oldText[end - 1] === newText[end - 1]) {
        end--
      }

      if (start < end) {
        const deletedText = oldText.slice(start, end)
        const insertedText = newText.slice(start, end)

        // Create a compound command that first deletes then inserts
        const deleteCommand = createTextDeleteCommand(
          wrappedSetText,
          () => oldText.slice(0, start) + oldText.slice(end),
          deletedText,
          start
        )

        const insertCommand = createTextInsertCommand(
          wrappedSetText,
          () => oldText.slice(0, start) + oldText.slice(end),
          insertedText,
          start
        )

        // Create a compound command that handles both operations
        const compoundCommand = {
          execute() {
            insertCommand.execute()
          },
          undo() {
            deleteCommand.undo()
          }
        }

        undoStack.addToStack(compoundCommand)
      }
    }

    updateCanUndoRedo()
  }

  const undo = () => {
    isUndoRedo = true
    const success = undoStack.undo()
    isUndoRedo = false
    if (success) {
      updateCanUndoRedo()
    }
  }

  const redo = () => {
    isUndoRedo = true
    const success = undoStack.redo()
    isUndoRedo = false
    if (success) {
      updateCanUndoRedo()
    }
  }

  const clear = () => {
    undoStack.clear()
    updateCanUndoRedo()
  }

  return {
    undo,
    redo,
    canUndo,
    canRedo,
    recordChange,
    clear
  }
}
