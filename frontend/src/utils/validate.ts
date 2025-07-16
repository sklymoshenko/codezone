// @ts-expect-error - window.ts is not typed
import libEsnext from 'lib.esnext.d.ts?raw'

export const validateTsCode = (code: string) => {
  const ts = window.ts
  if (!ts) {
    console.error('TypeScript compiler not loaded')
    return []
  }

  const fileName = 'file.ts'
  let sourceFile = ts.createSourceFile(fileName, '', ts.ScriptTarget.ESNext, true)

  const defaultLib = {
    fileName: 'lib.esnext.d.ts',
    content: libEsnext
  }

  const languageServiceHost = {
    getScriptFileNames: () => [fileName],
    getScriptVersion: () => '1',
    getScriptSnapshot: (filePath: string) => {
      if (filePath === fileName) {
        return ts.ScriptSnapshot.fromString(sourceFile.getFullText())
      }
      if (filePath === defaultLib.fileName) {
        return ts.ScriptSnapshot.fromString(defaultLib.content)
      }
      return undefined
    },
    getCurrentDirectory: () => '',
    getCompilationSettings: () => ({
      noEmit: true,
      strict: true,
      target: ts.ScriptTarget.ESNext,
      lib: ['esnext']
    }),
    getDefaultLibFileName: () => defaultLib.fileName,
    fileExists: (filePath: string) =>
      filePath === fileName || filePath === defaultLib.fileName,
    readFile: (filePath: string) => {
      if (filePath === fileName) return sourceFile.getFullText()
      if (filePath === defaultLib.fileName) return defaultLib.content
      return undefined
    }
  }

  const languageService = ts.createLanguageService(
    languageServiceHost,
    ts.createDocumentRegistry()
  )

  sourceFile = ts.createSourceFile(fileName, code, ts.ScriptTarget.ESNext, true)

  // Only include diagnostics with a file (filter out undefined)
  const syntactic = languageService
    .getSyntacticDiagnostics(fileName)
    .filter((d: any) => d.file)
  const semantic = (languageService.getSemanticDiagnostics(fileName) as any[]).filter(
    (d: any) => d.file
  )
  const diagnostics = syntactic.concat(semantic)

  const errors = diagnostics.filter((diagnostic: any) => {
    return (
      diagnostic.file &&
      diagnostic.file.fileName === fileName &&
      !diagnostic.messageText.toString().startsWith("Cannot find name 'console'")
    )
  })

  return errors.map((error: Record<string, any>) => {
    const { line, character } = error.file.getLineAndCharacterOfPosition(error.start)
    const message = ts.flattenDiagnosticMessageText(error.messageText, '\n')
    return `TypeScript Error [${error.code}]: ${message} on line ${line + 1}, character ${character + 1}`
  })
}
