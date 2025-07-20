# License Summary for CodeZone

## Overview

CodeZone is licensed under the **MIT License**, with proper attribution to all third-party dependencies.

## License Files Structure

- **`LICENSE`** - Main license file with full MIT license text and all third-party attributions
- **`NOTICE`** - Brief attribution notice for third-party components
- **`COPYRIGHT`** - Simple copyright statement
- **`frontend/LICENSE`** - Frontend-specific MIT license (consistent with main)
- **`LICENSE_SUMMARY.md`** - This summary document

## Key Dependencies & Their Licenses

### Core Dependencies

| Component                | License      | Copyright                | Usage                         |
| ------------------------ | ------------ | ------------------------ | ----------------------------- |
| **V8 JavaScript Engine** | BSD-3-Clause | Google Inc. & V8 authors | JavaScript execution via v8go |
| **v8go**                 | BSD-3-Clause | Roger Peppe              | Go bindings for V8            |
| **Goja**                 | MIT          | Dmitry Vyukov            | JavaScript engine for Windows |
| **Wails**                | MIT          | Lea Anthony              | Desktop app framework         |

### Frontend Dependencies

- **Primary**: MIT licensed packages (SolidJS, Vite, etc.)
- **Secondary**: ISC and Apache-2.0 licensed packages
- **Full list**: See `frontend/package.json`

### Go Dependencies

- **Listed in**: `go.mod` and `go.sum`
- **Primary licenses**: MIT, BSD-3-Clause, Apache-2.0

## Compliance Checklist ✅

- [x] **Main LICENSE file** with full MIT text
- [x] **V8 BSD-3-Clause attribution** included
- [x] **v8go BSD-3-Clause attribution** included
- [x] **Goja MIT attribution** included
- [x] **Wails MIT attribution** included
- [x] **Copyright notices** preserved for all dependencies
- [x] **Source file headers** added to key files
- [x] **NOTICE file** created for attribution
- [x] **Frontend license** aligned with main project

## Distribution Notes

When distributing CodeZone:

1. **Include all license files** (`LICENSE`, `NOTICE`, `COPYRIGHT`)
2. **Keep third-party attributions** intact
3. **Maintain copyright notices** in source files
4. **Reference dependency licenses** as documented

## Legal Safety ✅

- ✅ **Commercial use** permitted for all components
- ✅ **No copyleft restrictions** (no GPL dependencies)
- ✅ **Proper attribution** provided for all BSD-licensed components
- ✅ **Compatible licenses** (MIT + BSD-3-Clause + Apache-2.0)
- ✅ **No license conflicts**

---
