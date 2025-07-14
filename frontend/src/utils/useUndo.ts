import { createSignal, Accessor, Setter } from 'solid-js';
import { createUndoStack, createTextInsertCommand, createTextDeleteCommand } from './undoStack';

export interface UseUndoReturn {
  undo: () => void;
  redo: () => void;
  canUndo: Accessor<boolean>;
  canRedo: Accessor<boolean>;
  recordChange: (oldText: string, newText: string, cursorPos: number) => void;
  clear: () => void;
}

export function useUndo(
  text: Accessor<string>,
  setText: Setter<string>
): UseUndoReturn {
  const undoStack = createUndoStack();
  const [canUndo, setCanUndo] = createSignal(false);
  const [canRedo, setCanRedo] = createSignal(false);
  
  let isUndoRedo = false;
  
  const updateCanUndoRedo = () => {
    setCanUndo(undoStack.canUndo());
    setCanRedo(undoStack.canRedo());
  };

  const recordChange = (oldText: string, newText: string, cursorPos: number) => {
    if (isUndoRedo || oldText === newText) return;
    
    const oldLen = oldText.length;
    const newLen = newText.length;
    
    const wrappedSetText = (text: string) => {
      isUndoRedo = true;
      setText(text);
      isUndoRedo = false;
    };
    
    if (newLen > oldLen) {
      // Text was inserted
      const insertedText = newText.slice(cursorPos - (newLen - oldLen), cursorPos);
      const position = cursorPos - insertedText.length;
      
      const command = createTextInsertCommand(
        wrappedSetText,
        text,
        insertedText,
        position
      );
      
      undoStack.addToStack(command);
      
    } else if (newLen < oldLen) {
      // Text was deleted
      const deletedText = oldText.slice(cursorPos, cursorPos + (oldLen - newLen));
      
      const command = createTextDeleteCommand(
        wrappedSetText,
        () => newText,
        deletedText,
        cursorPos
      );
      
      undoStack.addToStack(command);
    }
    
    updateCanUndoRedo();
  };

  const undo = () => {
    isUndoRedo = true;
    const success = undoStack.undo();
    isUndoRedo = false;
    if (success) {
      updateCanUndoRedo();
    }
  };

  const redo = () => {
    isUndoRedo = true;
    const success = undoStack.redo();
    isUndoRedo = false;
    if (success) {
      updateCanUndoRedo();
    }
  };

  const clear = () => {
    undoStack.clear();
    updateCanUndoRedo();
  };

  return {
    undo,
    redo,
    canUndo,
    canRedo,
    recordChange,
    clear
  };
} 