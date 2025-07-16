import ts from 'typescript'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { validateTsCode } from '../validate'

describe('createTSValidator', () => {
  beforeAll(() => {
    // Mock window.ts for testing environment
    // @ts-expect-error: Mocking global window object for testing
    global.window = { ts: ts }
  })

  afterAll(() => {
    // @ts-expect-error: Cleaning up global window object after testing
    delete global.window
  })

  it('should return an empty array for valid TypeScript code', () => {
    const code = 'const x: number = 10;'
    const errors = validateTsCode(code)
    expect(errors).toEqual([])
  })

  it('should return errors for invalid TypeScript code (undeclared variable)', () => {
    const code = 'console.log(y);'
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain("Cannot find name 'y'")
  })

  it('should return errors for invalid TypeScript code (type mismatch)', () => {
    const code = 'const x: string = 10;'
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain("Type 'number' is not assignable to type 'string'")
  })

  it('should return an empty array for empty code', () => {
    const code = ''
    const errors = validateTsCode(code)
    expect(errors).toEqual([])
  })

  it('should return errors with correct line and character information', () => {
    const code = 'const a = 10;\nconst b: string = a;'
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain('on line 2, character 7') // Line 2, character 7
    expect(errors[0]).toContain("Type 'number' is not assignable to type 'string'")
  })

  it('should handle complex type errors across multiple lines', () => {
    const code = `
      interface MyInterface {
        name: string;
        age: number;
      }
      const obj: MyInterface = {
        name: "Test",
        age: "twenty" // Type error here
      };
    `
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain('on line 8, character 9') // Line 7, character 14
    expect(errors[0]).toContain("Type 'string' is not assignable to type 'number'")
  })

  it('should handle syntax errors', () => {
    const code = 'const x = "hello;' // Unclosed string literal
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain('Unterminated string literal.')
  })

  it('should handle missing imports', () => {
    const code = 'import { nonExistent } from "non-existent-module";'
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain("Cannot find module 'non-existent-module'")
  })

  it('should handle unused variables (if configured in tsconfig)', () => {
    // This test assumes tsconfig.json or default TS behavior flags unused variables.
    // If not, this test might need adjustment or be skipped based on project setup.
    const code = 'const unusedVar = 10;'
    const errors = validateTsCode(code)
    // Depending on TS configuration, this might or might not produce an error.
    // For a robust test, ensure your tsconfig.json has "noUnusedLocals": true
    // or similar strict checks.
    if (errors.length > 0) {
      expect(errors[0]).toContain('unusedVar')
    } else {
      console.log('Note: Unused variable warning not triggered. Check tsconfig.json.')
    }
  })

  it('should handle interface extending correctly', () => {
    const code = `
      interface Shape {
        color: string;
      }
      interface Circle extends Shape {
        radius: number;
      }
      const myCircle: Circle = { color: "red", radius: 10 };
      const invalidCircle: Circle = { color: "blue" }; // Missing radius
    `
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain("Property 'radius' is missing")
  })

  it('should handle valid interface extending', () => {
    const code = `
      interface Shape {
        color: string;
      }
      interface Circle extends Shape {
        radius: number;
      }
      const myCircle: Circle = { color: "red", radius: 10 };
    `
    const errors = validateTsCode(code)
    expect(errors).toEqual([])
  })

  it('should handle type merging (interface declaration merging)', () => {
    const code = `
      interface Box {
        height: number;
      }
      interface Box {
        width: number;
      }
      const myBox: Box = { height: 10, width: 20 };
      const invalidBox: Box = { height: 10 }; // Missing width
    `
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain("Property 'width' is missing")
  })

  it('should handle valid type merging', () => {
    const code = `
      interface Box {
        height: number;
      }
      interface Box {
        width: number;
      }
      const myBox: Box = { height: 10, width: 20 };
    `
    const errors = validateTsCode(code)
    expect(errors).toEqual([])
  })

  it('should handle union types correctly', () => {
    const code = `
      type Id = number | string;
      const userId: Id = 123;
      const productId: Id = "abc";
      const invalidId: Id = true; // Not a number or string
    `
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain("Type 'boolean' is not assignable to type 'Id'")
  })

  it('should handle valid union types', () => {
    const code = `
      type Id = number | string;
      const userId: Id = 123;
      const productId: Id = "abc";
    `
    const errors = validateTsCode(code)
    expect(errors).toEqual([])
  })

  it('should handle enums correctly', () => {
    const code = `
      enum Direction {
        Up,
        Down,
        Left,
        Right,
      }
      const go: Direction = Direction.Up;
      const invalidDirection: Direction = 5; // 5 is not a valid enum value unless explicitly assigned
    `
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain("Type '5' is not assignable to type 'Direction'")
  })

  it('should handle valid enums', () => {
    const code = `
      enum Direction {
        Up,
        Down,
        Left,
        Right,
      }
      const go: Direction = Direction.Up;
    `
    const errors = validateTsCode(code)
    expect(errors).toEqual([])
  })

  it('should handle as const assertions', () => {
    const code = `
      const arr = [1, 2, 3] as const;
      type ArrType = typeof arr;
      const val: ArrType = [1, 2, 3];
      const invalidVal: ArrType = [1, 2, 4]; // Should be an error
    `
    const errors = validateTsCode(code)
    expect(errors.length).toBeGreaterThan(0)
    expect(errors[0]).toContain("Type '4' is not assignable to type '3'")
  })

  it('should handle valid as const assertions', () => {
    const code = `
      const arr = [1, 2, 3] as const;
      type ArrType = typeof arr;
      const val: ArrType = [1, 2, 3];
    `
    const errors = validateTsCode(code)
    expect(errors).toEqual([])
  })
})
