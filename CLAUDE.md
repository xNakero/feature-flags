# Project Conventions

## Go Code Style

- **Descriptive interface parameter names**: use names that clarify intent beyond what the type already says (e.g. `flagValue domain.FlagValue` not `value domain.FlagValue`).
- **No obvious comments**: skip comments that restate basic Go knowledge (context propagation, error return patterns) or repeat what the method name already says. Only comment when the logic isn't self-evident.
- **Self-descriptive names need no comment**: error sentinels (e.g. `ErrNotFound`) and other identifiers whose names already convey full meaning must not have godoc comments.
- **Interface godocs describe what, not how**: a `// FlagStore is the outbound port for...` comment is correct. Error contracts and behavioral rules belong in code (tests, implementation), not interface comments.
- **No instructional comments in code**: comments that tell callers *how* to use something (e.g. "use errors.Is for comparison") belong in documentation (.md files), not in source.
- **No redundant godoc openers**: don't start a comment with a sentence that just restates the type name (e.g. `// CreateFlagRequest carries the data needed to create a new feature flag`).

## Architecture

- **Port interfaces use response DTOs**: service/port interfaces must never return domain objects. Define dedicated response DTOs in the port package even when they map field-for-field to a domain type.

## Feature Branch Scope

- **Respect ticket ownership**: before creating a file, verify it belongs to the current ticket. If a file is owned by a different ticket/branch, do not include it here — let the owning branch introduce it.
