Technology choice – Go
NEBOLON Runtime will be implemented in Go (Golang).

Go was selected because it provides:
- Standalone binaries – Runtime can be distributed without requiring users to install a separate runtime environment.
- Cross-platform support – Supports Windows, Linux and macOS.
- Performance – Fast execution with relatively low resource consumption.
- Built-in concurrency – Well suited for managing multiple block executions and asynchronous operations.
- Standard library – Strong built-in support for JSON, file handling, HTTP and process management.
- Simple deployment – Easy to build, distribute and execute as a standalone application.
The initial implementation will focus on a CLI-based Runtime, but the internal architecture should allow the same execution engine to be used later by NEBOLON Studio, web applications and other interfaces.

Important architectural principle: The Runtime should be designed independently of its interface. The CLI is simply the first way to interact with it.