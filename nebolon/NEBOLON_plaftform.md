# NEBOLON – Platform Name and Naming Hierarchy

## Overview

**NEBOLON** is a strong candidate for the name of the entire visual, modular programming platform and its ecosystem.

The name is intended to serve as an umbrella brand for multiple components: Studio, Runtime, CLI, block libraries, and future tools. It therefore does not need to describe a single application or a specific programming format.

**Proposed positioning:**

**NEBOLON** – a modular visual programming ecosystem.

## Proposed Naming Hierarchy

| Component | Proposed Name | Responsibility |
|------------|---------------|----------------|
| Entire platform | **NEBOLON** | Product and ecosystem umbrella brand |
| Visual development environment | **NEBOLON Studio** | Build, organize, and edit visual programs |
| Execution engine | **NEBOLON Runtime** | Read Flow Maps and execute programs and blocks |
| Block collection | **NEBOLON Blocks** | Collective name for the platform's block ecosystem |
| Block catalog | **NEBOLON Block Library** | Repository of block definitions and implementations |
| Command-line tool | **NEBOLON CLI** | Execute and manage flows from the terminal |
| Block editor | **NEBOLON Block Editor** | Configure and edit blocks |

## File Formats

There is no compelling reason to change the working names of the file formats simply because the platform receives its own brand identity.

| Format | Proposed Extension | Content |
|----------|-------------------|---------|
| Flow Map | `.flow` | Describes the program canvas, nodes, positions, connections, and settings |
| Block Definition | `.block` | Describes a block's identity, category, ports, types, configuration, and implementation |

A Flow Map may reference block definitions as well as other Flow Maps, such as sub-flows.

**Core principle:** NEBOLON is the platform. `.flow` and `.block` are formats used by the platform.

## Usage Examples

### In the Terminal

```bash
nebolon execute --flow hello.flow --node hello1
```

### In a Project

```text
my-project/
├── main.flow
├── blocks/
│   ├── http-request.block
│   └── process-data.block
└── subflows/
    └── data-processing.flow
```

### In Documentation and User Interfaces

- Create a Block in NEBOLON.
- Run your Flow with NEBOLON Runtime.
- Install blocks from the NEBOLON Block Library.
- Build your first Flow in NEBOLON Studio.

## Block Library Naming

Two closely related concepts can be used with distinct meanings:

- **NEBOLON Blocks** is the overarching name for the block ecosystem and available block offerings.
- **NEBOLON Block Library** is the concrete repository that contains block definitions and their associated implementations.

Example:

> Install blocks from the NEBOLON Block Library.

## Principles to Maintain

- **One umbrella brand:** NEBOLON brings together the platform's various components.
- **Clear component names:** Studio, Runtime, CLI, and Block Library clearly describe each component's role.
- **Independent format names:** `.flow` and `.block` can continue to exist independently of the platform brand.
- **Platform-independent architecture:** The same Flow Map and block model should be usable across different environments and interfaces.
- **Separation of responsibilities:** Flow Maps describe programs, blocks provide functionality, the Runtime handles execution, and renderers are responsible for presentation.

## Status and Next Steps

**NEBOLON is a promising naming candidate, not a final or legally verified name selection.**

A preliminary search has not revealed any clearly established software platform using the exact name **NEBOLON**. However, similar names have appeared, including **EMPYTRIUM** and **emtrion**. As a result, similarity and potential confusion risks should be investigated further.

The following checks remain to be completed:

- Verify domain availability for `nebolon.com`, `nebolon.se`, `nebolon.io`, `nebolon.dev`, and `nebolon.app` through a domain registrar.
- Search for exact matches and similar spellings in PRV, EUIPO, TMview, and WIPO trademark databases.
- Review relevant trademark classes covering software, development tools, and technical services.
- Check usage on software platforms and package registries, such as GitHub and npm.

Relevant resources:

- PRV Databases: https://www.prv.se/sv/databaser/
- EUIPO eSearch: https://euipo.europa.eu/eSearch/
- TMview: https://www.tmdn.org/tmview/
- WIPO Global Brand Database: https://branddb.wipo.int/

This document summarizes a conceptual naming proposal and a proposed naming hierarchy. It does not constitute a legal trademark assessment.
