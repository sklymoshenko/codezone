export interface UndoCommand {
  execute(): void;
  undo(): void;
  merge?(other: UndoCommand): boolean;
}

export function createTextInsertCommand(
  setText: (text: string) => void,
  getText: () => string,
  insertedText: string,
  position: number,
  timestamp: number = Date.now()
): UndoCommand {
  let currentInsertedText = insertedText;
  let currentTimestamp = timestamp;

  return {
    execute() {
      const currentText = getText();
      const newText = currentText.slice(0, position) + currentInsertedText + currentText.slice(position);
      setText(newText);
    },

    undo() {
      const currentText = getText();
      const newText = currentText.slice(0, position) + currentText.slice(position + currentInsertedText.length);
      setText(newText);
    },

    merge(other: UndoCommand): boolean {
      if (!('insertedText' in other) || !('timestamp' in other)) return false;
      
      const otherCmd = other as any;
      const timeDiff = otherCmd.timestamp - currentTimestamp;
      const isConsecutive = otherCmd.position === position + currentInsertedText.length;
      const isSingleChar = otherCmd.insertedText.length === 1 && currentInsertedText.length < 50;
      
      if (timeDiff < 1000 && isConsecutive && isSingleChar) {
        currentInsertedText += otherCmd.insertedText;
        currentTimestamp = otherCmd.timestamp;
        return true;
      }
      
      return false;
    }
  };
}

export function createTextDeleteCommand(
  setText: (text: string) => void,
  getText: () => string,
  deletedText: string,
  position: number,
  timestamp: number = Date.now()
): UndoCommand {
  let currentDeletedText = deletedText;
  let currentPosition = position;
  let currentTimestamp = timestamp;

  return {
    execute() {
      const currentText = getText();
      const newText = currentText.slice(0, currentPosition) + currentText.slice(currentPosition + currentDeletedText.length);
      setText(newText);
    },

    undo() {
      const currentText = getText();
      const newText = currentText.slice(0, currentPosition) + currentDeletedText + currentText.slice(currentPosition);
      setText(newText);
    },

    merge(other: UndoCommand): boolean {
      if (!('deletedText' in other) || !('timestamp' in other)) return false;
      
      const otherCmd = other as any;
      const timeDiff = otherCmd.timestamp - currentTimestamp;
      const isBackspace = otherCmd.position === currentPosition - otherCmd.deletedText.length;
      const isSingleChar = otherCmd.deletedText.length === 1 && currentDeletedText.length < 50;
      
      if (timeDiff < 1000 && isBackspace && isSingleChar) {
        currentDeletedText = otherCmd.deletedText + currentDeletedText;
        currentPosition = otherCmd.position;
        currentTimestamp = otherCmd.timestamp;
        return true;
      }
      
      return false;
    }
  };
}

export function createUndoStack(maxSize: number = 100) {
  const undoStack: UndoCommand[] = [];
  const redoStack: UndoCommand[] = [];

  return {
    execute(command: UndoCommand): void {
      // Try to merge with the last command
      const lastCommand = undoStack[undoStack.length - 1];
      if (lastCommand && lastCommand.merge?.(command)) {
        return;
      }

      // Execute the command
      command.execute();
      
      // Add to undo stack
      undoStack.push(command);
      
      // Clear redo stack
      redoStack.length = 0;
      
      // Limit stack size
      if (undoStack.length > maxSize) {
        undoStack.shift();
      }
    },

    undo(): boolean {
      if (undoStack.length === 0) return false;
      
      const command = undoStack.pop()!;
      command.undo();
      redoStack.push(command);
      
      return true;
    },

    redo(): boolean {
      if (redoStack.length === 0) return false;
      
      const command = redoStack.pop()!;
      command.execute();
      undoStack.push(command);
      
      return true;
    },

    canUndo(): boolean {
      return undoStack.length > 0;
    },

    canRedo(): boolean {
      return redoStack.length > 0;
    },

    clear(): void {
      undoStack.length = 0;
      redoStack.length = 0;
    },

    // For bypassing execute when text is already changed
    addToStack(command: UndoCommand): void {
      const lastCommand = undoStack[undoStack.length - 1];
      if (lastCommand && lastCommand.merge?.(command)) {
        return;
      }
      
      undoStack.push(command);
      redoStack.length = 0;
      
      if (undoStack.length > maxSize) {
        undoStack.shift();
      }
    }
  };
} 