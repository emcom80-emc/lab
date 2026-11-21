# NEBOLON Runtime – First Milestone: Simulated Block Execution

## Project context

We are building **NEBOLON**, a platform-independent visual programming ecosystem where programs are created by connecting functional blocks on a canvas.

The platform is designed around four separate responsibilities:

- **Flow Maps** describe program structure and connections.
- **Block Definitions** describe blocks, their interfaces and implementation references.
- **Runtime** handles program execution.
- **Renderers** handle how programs and blocks are presented in different environments.

These responsibilities must remain independent.

The first task is to implement the initial version of **NEBOLON Runtime**.

## Technology choice

The Runtime must be implemented in **Go (Golang)**.

Reasons for choosing Go:

- Standalone executable binaries.
- Cross-platform support (Windows, Linux and macOS).
- Good performance and relatively low resource consumption.
- Built-in concurrency support.
- Strong standard library support for JSON, files and process management.
- Simple deployment without requiring users to install a separate runtime environment.

Avoid unnecessary external dependencies. Prefer the Go standard library whenever practical.

The initial interface will be a CLI, but the internal execution engine must be designed so it can later be used by other interfaces, such as NEBOLON Studio, web applications and desktop applications.

The CLI must not contain the core execution logic. It should act as an interface to the Runtime.

## First milestone: Simulated Block Execution

The purpose of this milestone is to establish and test the basic execution pipeline without implementing communication with external block processes.

**Do not implement actual block process execution yet.**

The Runtime must be able to:

1. Receive a Flow Map and an execution command.
2. Read the Flow Map from a JSON file.
3. Identify the requested block by its node ID.
4. Read the corresponding Block Definition from the block library.
5. Locate and inspect the block's implementation reference.
6. Attempt to invoke the implementation through a clearly separated execution interface.
7. Since the block execution API is not implemented yet, return a simulated, standardized failure result.
8. Update the selected block's execution status and result in the Flow Map.
9. Return the updated Flow Map as JSON.

The simulated failure is intentional. It allows us to build and test the execution pipeline before implementing actual block communication.

## Standardized execution result

When block execution cannot be performed because the block execution API is not implemented, the Runtime should produce a result equivalent to:

```json
{
  "id": "hello1",
  "type": "hello-world",
  "status": "failed",
  "outputs": {},
  "success": false,
  "fail": true,
  "error": {
    "code": "BLOCK_API_MISSING",
    "message": "Block execution API is not implemented yet."
  }
}
```

The Runtime must use structured error codes and messages rather than relying on plain text errors.

This simulated error belongs to the Runtime's execution layer, not permanently in the Block Definition.

When actual block execution is implemented later, the simulated failure should be replaceable without requiring a redesign of the surrounding execution pipeline.

## Execution flow semantics

Regular blocks use the following standardized execution ports:

- `exec`: execution input.
- `success`: execution output, activated when execution succeeds.
- `fail`: execution output, activated when execution fails.

Data outputs and error data are separate from execution signals.

For example:

- `outputs.data` contains regular output data.
- `error` contains error information.
- `success` and `fail` represent execution flow signals.

For this first milestone, only the requested block needs to be processed. Do not implement traversal or execution of connected blocks yet.

## CLI interface

Implement an initial CLI command following this general pattern:

```bash
nebolon execute --flow hello.flow --node hello1 --output hello.result.flow
```

The command should:

- Read the input Flow Map.
- Execute the requested node.
- Write the updated Flow Map to the specified output file.
- Never overwrite the original input file automatically.
- Return machine-readable JSON to stdout when no output file is specified.
- Write diagnostics and error messages to stderr.

Keep the CLI interface separate from the execution engine.

## File and data handling

Flow Maps and Block Definitions are JSON-based formats.

The Runtime should work with Flow Map and Block Definition objects in memory. File reading and writing should be handled at the appropriate application/CLI boundary.

Do not assume that file access will always be used in the future. Other applications must eventually be able to pass these objects directly to the Runtime through an API.

If the exact JSON structures or directory conventions have not yet been established in the project, inspect the existing files first. If necessary, propose a minimal initial structure, document it and keep it easy to evolve.

Do not silently introduce incompatible assumptions.

## Architecture and code quality

Keep the implementation modular and maintainable.

In particular:

- Separate CLI handling from Runtime execution logic.
- Separate Flow Map parsing from execution logic.
- Separate Block Definition loading from execution logic.
- Keep the block invocation mechanism behind a clear interface.
- Make it possible to replace the simulated execution with real block execution later.
- Use structured errors.
- Handle invalid JSON, missing files, missing nodes, missing block definitions and invalid input gracefully.
- Add appropriate tests for the execution pipeline and error cases.

Do not implement features outside the scope of this milestone.

## Important: Future architecture

The following capabilities are planned for later development, but must NOT be implemented in this milestone:

- Starting external block processes.
- Communication with blocks through stdin/stdout.
- JSON Lines block communication protocol.
- Python, JavaScript or other language adapters.
- Execution of entire connected flows.
- Asynchronous block execution.
- Cancellation and parallel execution.
- Runtime server or persistent runtime process.
- User interface or visual editor.

However, the current architecture must not make these future capabilities difficult to add.

In particular, actual block execution should eventually be possible through isolated processes and a language-independent protocol. Avoid tightly coupling the Runtime to any particular block implementation language.

## Development approach

Before implementing anything:

1. Inspect the existing project directory and files.
2. Identify what already exists and what is missing.
3. Propose a simple project structure for this first milestone.
4. Explain the main architectural decisions briefly.
5. Identify any assumptions about Flow Map and Block Definition formats.

Then implement the milestone.

Do not overwrite or delete existing project files without first explaining why.

Keep the implementation focused. Do not over-engineer or attempt to build the complete NEBOLON platform.

## Acceptance criteria

The milestone is complete when:

- The project builds successfully with Go.
- The CLI can load a Flow Map from a JSON file.
- The Runtime can identify a requested node.
- The Runtime can load its corresponding Block Definition.
- The Runtime attempts to resolve the block implementation through the execution interface.
- The Runtime returns the standardized `BLOCK_API_MISSING` result.
- The selected node's status and result are updated in the returned Flow Map.
- The updated Flow Map can be written to a separate output file.
- The original Flow Map remains unchanged.
- Errors are handled consistently.
- Automated tests cover the main execution path and important failure cases.

**Start by inspecting the existing project and presenting your proposed implementation plan. Do not begin coding until the plan and any important assumptions have been presented.**