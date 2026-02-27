# Project Conventions

## Go Code Style

- **Descriptive interface parameter names**: use names that clarify intent beyond what the type already says (e.g. `flagValue domain.FlagValue` not `value domain.FlagValue`).
- **No obvious comments**: skip comments that restate basic Go knowledge (context propagation, error return patterns) or repeat what the method name already says. Only comment when the logic isn't self-evident.
- **Interface godocs describe what, not how**: a `// FlagStore is the outbound port for...` comment is correct. Error contracts and behavioral rules belong in code (tests, implementation), not interface comments.
