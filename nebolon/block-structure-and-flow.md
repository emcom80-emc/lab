Flow -- Block Structure and Execution Flow
1. Overview
Flow uses a standardized block structure and a unified execution-flow
model.

Every regular block follows the same basic execution interface,
regardless of its purpose, implementation language, or internal
functionality.

The core principles are:
- Blocks receive execution signals through a standard exec input.
- Blocks perform their operation when triggered.
- Blocks expose standardized execution outputs indicating the outcome.
- Execution outputs can be connected directly to other blocks'
  execution inputs.
- Data outputs are separate from execution signals.
This provides a consistent way to build programs visually without
requiring special execution logic for individual block types.
2. Standard Block Structure
Every regular block has the following conceptual structure:
┌──────────────────────────────┐
│             Block            │
│                              │
│  Execution Input             │
│  ┌──────────┐                │
│  │   exec   │                │
│  └──────────┘                │
│                              │
│  Execution Outputs           │
│  ┌──────────┐ ┌──────────┐    │
│  │ success  │ │   fail   │    │
│  └──────────┘ └──────────┘    │
│                              │
│  Data Outputs                │
│  ┌────────────────────────┐  │
│  │ outputs                │  │
│  └────────────────────────┘  │
│                              │
│  Error Data                  │
│  ┌────────────────────────┐  │
│  │ error                  │  │
│  └────────────────────────┘  │
└──────────────────────────────┘
Execution input: exec
- The standard execution input of a regular block.
- Receiving an execution signal starts the block's operation.
- All regular blocks use the same input name and concept.
- A block can have additional data inputs that provide values required
  for its operation.
Execution outputs
success
- Activated when the block completes successfully.
- Can be connected directly to another block's exec input.
- Allows execution to continue along the success path.
fail
- Activated when the block fails.
- Can be connected directly to another block's exec input.
- Allows execution to continue along an error-handling path.
These are standardized execution outputs available to all regular
blocks.
Data outputs
outputs
- Contains the block's regular output values.
- Can be connected to data inputs on other blocks.
- Is independent of execution signals.
error
- Contains an error message or structured error information when
  execution fails.
- Can be connected to other blocks as regular data.
- Can be used for logging, debugging, notifications, or error
  handling.
3. Execution Flow and Connections
Flow separates execution connections from data connections.
- Execution connections determine when blocks run.
- Data connections determine which values are passed between
  blocks.
For example:
             ┌─────────────┐
             │  Fetch Data │
             └──────┬──────┘
                    │ exec
                    ▼
             ┌─────────────┐
             │ Process Data│
             └──────┬──────┘
                    │
             ┌──────┴──────┐
             ▼             ▼
          success         fail
             │             │
             ▼             ▼
       Save Result     Handle Error
A more concrete example:
FetchData.success ───────► ProcessData.exec

FetchData.fail ──────────► HandleError.exec

FetchData.outputs.data ──► ProcessData.inputs.data

FetchData.error ─────────► HandleError.inputs.message
The execution connection determines when the next block starts, while
the data connection provides the information it needs.

Execution and data can therefore be connected independently.
4. Triggers
Triggers are blocks or external events that initiate execution.

Unlike regular blocks, trigger blocks do not necessarily require an
exec input. Their purpose is to generate execution signals when a
particular event occurs.

Examples include:
- Manual execution initiated by the runtime.
- Timers, such as Every Hour.
- Conditions becoming true.
- Incoming HTTP requests.
- External events or messages.
- Scheduled tasks.
A trigger generates an execution signal that can be connected to a
regular block's exec input.
Example: Timer trigger
┌───────────────┐
│  Every Hour   │
└───────┬───────┘
        │ trigger
        ▼
┌───────────────┐
│  Fetch Data   │
└───────┬───────┘
        │ success
        ▼
┌───────────────┐
│ Process Data  │
└───────────────┘
The timer generates a trigger every hour. The trigger starts
Fetch Data, which then continues through its success path.

The runtime can also initiate execution directly, without requiring a
trigger block.
5. Boolean Conditions
Condition blocks can use boolean outputs to determine which execution
path should be activated.

For example:
             ┌─────────────┐
             │  Condition  │
             └──────┬──────┘
                    │
             ┌──────┴──────┐
             ▼             ▼
           true          false
             │             │
             ▼             ▼
        Block A.exec  Block B.exec
A true or false output represents a condition result, not
necessarily a completed block execution.

This distinction is important:

  Connection / port   Purpose
  exec              Execution input
success           Successful execution output
fail              Failed execution output
true / false    Boolean condition outputs
outputs.*         Regular data outputs
error             Error information

A boolean output can be used by a condition or trigger to decide which
execution path to activate. When it is used to start a regular block, it
connects to that block's exec input.
6. Error Handling
Error handling is part of the standardized block interface rather than
something every block must implement differently.

When a block fails:
1. The block's execution ends.
2. Its fail execution output is activated.
3. Error information is made available through the error data output.
4. Connected error-handling blocks can be started.
5. The runtime records the failure and updates the block's execution
   status.
For example:
             ┌─────────────┐
             │ API Request │
             └──────┬──────┘
                    │
             ┌──────┴──────┐
             │             │
          success         fail
             │             │
             ▼             ▼
       Process Data    Log Error
                           ▲
                           │ error data
The runtime should also retain error information for debugging and
inspection, independently of whether an error-handling path is
connected.
7. Runtime Responsibilities
The runtime is responsible for managing execution, not individual block
visualizations.

Its responsibilities include:
- Receiving a Flow Map and an execution command.
- Loading and interpreting block definitions from the block-definition
  library.
- Resolving block types and their available inputs and outputs.
- Validating connections and compatible types.
- Determining which blocks should execute.
- Delivering execution signals to blocks.
- Managing data connections and passing values between blocks.
- Handling success and failure paths.
- Recording execution status and error information.
- Returning an updated Flow Map containing execution results.
For the initial version, the runtime can execute blocks internally
without introducing the separate block-process communication protocol.
Its architecture should leave room for that isolation layer later.
8. Important Architectural Principles
1. Standardized execution interface: All regular blocks use the
   same exec input and success / fail execution outputs.
2. Separate execution and data flow: Execution connections control
   when blocks run; data connections transfer values.
3. Triggers are execution sources: Timers, conditions, manual
   commands, and external events can initiate execution using the same
   execution mechanism.
4. Standardized error handling: All blocks expose failure through
   the same mechanism, with error information available as data.
5. Runtime-independent block definitions: The runtime reads block
   definitions from a library and uses them to understand block
   capabilities and connections.
6. Consistent behavior across implementations: A block's language
   or implementation method does not change its external execution
   interface.
7. Runtime owns execution state: Block status, execution results,
   and errors are managed by the runtime and represented in the updated
   Flow Map.
9. Conceptual Summary
Execution sources
(Manual / Timer / Condition / External Event)
                    │
                    ▼
             Regular Block
               exec input
                    │
             ┌──────┴──────┐
             ▼             ▼
          success         fail
             │             │
             ▼             ▼
       Next block     Error handler

Data outputs and error information
are passed separately through data connections.
The fundamental idea: A block performs an operation, reports whether
it succeeded or failed, and exposes its resulting data. Its execution
outputs determine what happens next. The runtime manages the entire
process, while the Flow Map describes the structure and connections.