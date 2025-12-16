# LLM Context Files for Chimborazo

Documentation optimized for local LLM consumption. These files help models like Qwen and DeepSeek understand the codebase without scanning all source files.

## Files

| File | Purpose | When to Use |
|------|---------|-------------|
| `TYPES.md` | All type definitions | When implementing any feature |
| `INTERFACES.md` | Interface contracts | When implementing a module |
| `PATTERNS.md` | Code patterns to follow | Always include |
| `OPERATIONS.md` | Geometry operation specs | When working on geometry |

## Usage

Include relevant files in your prompt:

```
<context>
[Contents of TYPES.md]
[Contents of PATTERNS.md]
</context>

Implement the HTTP fetcher according to the spec in specs/001-http-fetcher.md
```

## Token Estimates

| File | ~Tokens |
|------|---------|
| TYPES.md | 800 |
| INTERFACES.md | 600 |
| PATTERNS.md | 1000 |
| OPERATIONS.md | 1500 |
| **Total** | **~4000** |

Fits easily in 8K-32K context windows.
