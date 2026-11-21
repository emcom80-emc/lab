Project Overview – Flow

1. Core Concept and Vision
A platform-independent visual programming environment where programs are built by connecting functional blocks on a canvas. The goal is to run the same programs in web browsers, desktop applications, terminals, and other environments.

2. Flow Maps (.flow)
A JSON-based format describing canvases, nodes, positions, connections, configuration, and execution state. Flow Maps are the source of truth and can be edited visually or directly as text.

3. Block Definitions (.block)
JSON-based definitions describing each block's identity, inputs, outputs, data types, and implementation references. All blocks follow the same underlying model, from simple operations to advanced AI agents.

4. Flow Runtime
A standalone runtime written in Go, responsible for loading Flow Maps, managing execution, dataflow, control flow, and block lifecycles. It should support CLI execution initially, with long-running process and server modes later.

5. Isolated Block Communication and Protocol
The runtime communicates with blocks through a versioned, language-independent JSON protocol. The initial transport will use separate processes with JSON Lines over stdin/stdout, allowing blocks to be implemented in different programming languages.

6. Execution Model
The system combines dataflow (values flowing between blocks) and control flow (execution signals between blocks). The runtime must support executing individual nodes as well as entire flows.

7. Runtime State and Results
The runtime receives a Flow Map and returns an updated Flow Map containing statuses, outputs, and execution results. The runtime can maintain the loaded Flow Map in memory, while the initial CLI implementation will start a new process for each invocation.

8. Block Visualization
Each block's appearance is described using an internal, limited, declarative UI format. The runtime remains unaware of visualization details; different renderers interpret the same definition for web, terminal, desktop, and other platforms.

9. Editor and Configuration
The platform initially provides generic editors based on each block's declared configuration and data types. The architecture remains open to block-specific editors in the future.

10. Canvas and User Interface
A visual environment for creating and navigating flows, selecting blocks, and editing their configuration. The canvas owns positioning, connections, selection, and outer block frames, while blocks provide their own internal visual content.

11. Agents, Sub-canvases, and Security
Agent blocks should be able to create and operate within explicitly assigned sub-canvases without controlling the user's root canvas. Permissions, sandboxing, resource limits, and isolation are enforced by the runtime and its APIs.

12. Development Strategy and First Milestone
Development will proceed incrementally, validating the architecture from the beginning rather than building the entire interface first. The first milestone is a CLI-based Hello World: load a .flow file, start an isolated block process, receive its response, and return an updated Flow Map.

Overall architectural principle: Flow Maps describe programs, blocks provide functionality, the runtime handles execution, and renderers handle presentation. These responsibilities should remain as independent as possible.