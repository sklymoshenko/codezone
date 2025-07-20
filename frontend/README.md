# Code-executor

Code-executor is a powerful and intuitive desktop application for running code snippets in various languages. It provides a simple and clean interface for writing and executing code, with support for syntax highlighting and more.

## Features

- **Multi-language Support**: Write and execute code in JavaScript, Go, and SQL.
- **Syntax Highlighting**: Real-time syntax highlighting for improved readability.
- **Cross-Platform**: Runs on Windows, macOS, and Linux.
- **Lightweight**: Built with Wails to be a lightweight and performant application.

## Tech Stack

-   **[SolidJS](https://www.solidjs.com/)**: Frontend framework
-   **[Wails](https://wails.io/)**: Backend and windowing
-   **[TailwindCSS](https://tailwindcss.com/)**: Styling
-   **[Highlight.js](https://highlightjs.org/)**: Syntax highlighting
-   **[Bun](https://bun.sh/)**: JavaScript runtime and toolkit
-   **[Goja](https://github.com/dop251/goja)**: JavaScript engine for Windows
-   **[V8](https://github.com/v8/v8)**: JavaScript engine for Unix systems

## Development

To run the application in development mode, you need to have Go and Bun installed.

Then, run the following command:

```bash
wails dev
```

## Building

To build the application for your platform, run:

```bash
wails build
```

This will create a production-ready executable in the `build/bin` directory.
