# LSIF indexers

Language support is an ever-evolving feature of Sourcegraph. Some languages may be better supported than others due to demand or developer bandwidth/expertise. This page clarifies the status of the LSIF indexers which the Sourcegraph team can both recommend to customers and provide support for.

For a more complete description of the LSIF indexer ecosystem, see [LSIF.dev](https://lsif.dev/). A Sourcegraph instance can ingest any LSIF index file confirming to the [LSIF specification](https://microsoft.github.io/language-server-protocol/specifications/lsif/0.4.0/specification/). The absence of a third-party indexer on this page is not a quality judgment on that indexer; it is that Sourcegraph engineers may not have the required knowledge to provide deep technical support.

## Quick reference

This table is maintained as an authoritative resource for users, Sales, and Customer Engineers. Any major changes to the development of our indexers will be reflected here.

<table>
   <thead>
      <tr>
        <th>Indexer</th>
        <th>Status</th>
        <th><a href="#m1-can-provide-a-decorated-ast">M1:<br/><small>AST</small></a></th>
        <th><a href="#m2-emits-documents-and-ranges">M2:<br/><small>Docs / ranges</small></a></th>
        <th><a href="#m3-emits-hover-text">M3:<br/><small>Hover text</small></a></th>
        <th><a href="#m4-emits-definitions-within-compilation-unit">M4:<br/><small>Unit defs</small></a></th>
        <th><a href="#m5-emits-references-within-compilation-unit">M5:<br/><small>Unit refs</small></a></th>
        <th><a href="#m6-emits-definitions-within-input-source">M6:<br/><small>Input defs</small></a></th>
        <th><a href="#m7-emits-references-within-input-source">M7:<br/><small>Input refs</small></a></th>
        <th><a href="#m8-emits-monikers-for-cross-repository-support">M8:<br/><small>Cross-repo</small></a></th>
      </tr>
   </thead>
   <tbody>
      <tr>
        <td><a href="https://github.com/sourcegraph/lsif-go">lsif-go</a></td>
        <td><img src="https://img.shields.io/badge/status-ready-green" alt="Ready"></td>
        <td class="indexer-implemented-y">✓</td> <!-- M1 -->
        <td class="indexer-implemented-y">✓</td> <!-- M2 -->
        <td class="indexer-implemented-y">✓</td> <!-- M3 -->
        <td class="indexer-implemented-y">✓</td> <!-- M4 -->
        <td class="indexer-implemented-y">✓</td> <!-- M5 -->
        <td class="indexer-implemented-y">✓</td> <!-- M6 -->
        <td class="indexer-implemented-y">✓</td> <!-- M7 -->
        <td class="indexer-implemented-y">✓</td> <!-- M8 -->
      </tr>
      <tr>
        <td><a href="https://github.com/sourcegraph/lsif-node">lsif-node</a></td>
        <td><img src="https://img.shields.io/badge/status-ready-green" alt="Ready"></td>
        <td class="indexer-implemented-y">✓</td> <!-- M1 -->
        <td class="indexer-implemented-y">✓</td> <!-- M2 -->
        <td class="indexer-implemented-y">✓</td> <!-- M3 -->
        <td class="indexer-implemented-y">✓</td> <!-- M4 -->
        <td class="indexer-implemented-y">✓</td> <!-- M5 -->
        <td class="indexer-implemented-y">✓</td> <!-- M6 -->
        <td class="indexer-implemented-y">✓</td> <!-- M7 -->
        <td class="indexer-implemented-y">✓</td> <!-- M8 -->
      </tr>
      <tr>
        <td><a href="https://github.com/sourcegraph/lsif-clang">lsif-clang</a></td>
        <td><img src="https://img.shields.io/badge/status-ready-green" alt="Ready"></td>
        <td class="indexer-implemented-y">✓</td> <!-- M1 -->
        <td class="indexer-implemented-y">✓</td> <!-- M2 -->
        <td class="indexer-implemented-y">✓</td> <!-- M3 -->
        <td class="indexer-implemented-y">✓</td> <!-- M4 -->
        <td class="indexer-implemented-y">✓</td> <!-- M5 -->
        <td class="indexer-implemented-y">✓</td> <!-- M6 -->
        <td class="indexer-implemented-y">✓</td> <!-- M7 -->
        <td class="indexer-implemented-n">✗</td> <!-- M8 -->
      </tr>
      <tr>
        <td><a href="https://github.com/sourcegraph/lsif-java">lsif-java</a></td>
        <td><img src="https://img.shields.io/badge/status-development-white" alt="Development"></td>
        <td class="indexer-implemented-y">✓</td> <!-- M1 -->
        <td class="indexer-implemented-y">✓</td> <!-- M2 -->
        <td class="indexer-implemented-y">✓</td> <!-- M3 -->
        <td class="indexer-implemented-y">✓</td> <!-- M4 -->
        <td class="indexer-implemented-y">✓</td> <!-- M5 -->
        <td class="indexer-implemented-y">✓</td> <!-- M6 -->
        <td class="indexer-implemented-y">✓</td> <!-- M7 -->
        <td class="indexer-implemented-n">✗</td> <!-- M8 -->
      </tr>
      <tr>
        <td><a href="https://github.com/sourcegraph/lsif-semanticdb">lsif-semanticdb</a></td>
        <td><img src="https://img.shields.io/badge/status-beta-yellow" alt="Beta"></td>
        <td class="indexer-implemented-y">✓</td> <!-- M1 -->
        <td class="indexer-implemented-y">✓</td> <!-- M2 -->
        <td class="indexer-implemented-y">✓</td> <!-- M3 -->
        <td class="indexer-implemented-y">✓</td> <!-- M4 -->
        <td class="indexer-implemented-y">✓</td> <!-- M5 -->
        <td class="indexer-implemented-y">✓</td> <!-- M6 -->
        <td class="indexer-implemented-y">✓</td> <!-- M7 -->
        <td class="indexer-implemented-n">✗</td> <!-- M8 -->
      </tr>
      <tr>
        <td><a href="https://github.com/sourcegraph/lsif-cpp">lsif-cpp</a></td>
        <td><img src="https://img.shields.io/badge/status-deprecated-red" alt="Deprecated"></td>
        <td class="indexer-implemented-y">✓</td> <!-- M1 -->
        <td class="indexer-implemented-y">✓</td> <!-- M2 -->
        <td class="indexer-implemented-n">✗</td> <!-- M3 -->
        <td class="indexer-implemented-y">✓</td> <!-- M4 -->
        <td class="indexer-implemented-y">✓</td> <!-- M5 -->
        <td class="indexer-implemented-y">✓</td> <!-- M6 -->
        <td class="indexer-implemented-y">✓</td> <!-- M7 -->
        <td class="indexer-implemented-n">✗</td> <!-- M8 -->
      </tr>
   </tbody>
</table>

### Legend

- **M1**: Can provide a decorated AST
- **M2**: Emits documents and ranges
- **M3**: Emits hover text
- **M4**: Emits definitions (within compilation unit)
- **M5**: Emits references (within compilation unit)
- **M6**: Emits definitions (within input source)
- **M7**: Emits references (within input source)
- **M8**: Emits monikers for cross-repository support

#### Status definitions

An indexer status is:

- <img src="https://img.shields.io/badge/status-ready-green" alt="Ready"> (_ready_): When no major features are absent but edge cases remain.
- <img src="https://img.shields.io/badge/status-beta-yellow" alt="Beta"> (_beta_): When it could be useful to early adopters despite lack of features.
- <img src="https://img.shields.io/badge/status-alpha-orange" alt="Alpha"> (_alpha_): When early adopters can try it with expectations of failure.
- <img src="https://img.shields.io/badge/status-development-white" alt="Development"> (_development_): When we are actively working on implementation.
- <img src="https://img.shields.io/badge/status-deprecated-red" alt="Deprecated"> (_deprecated_): When we are no longer maintaining a solution. In these cases we will have a migration path to an alternate indexer.

## Milestone definitions

A common set of steps required to build feature-complete LSIF indexers is broadly outlined below. The implementation order and _doneness criteria_ of these steps may differ between language and development ecosystems. Major divergences will be detailed in the notes below.

### Basic functionality

#### M1: Can provide a decorated AST

The indexer can read produce an in-memory abstract syntax tree from source code input. Source code may be read from disk independently of a particular build system. This syntax tree should be decorated with symbol and type information (if supported). This is likely the step that will take the longest as the choice of compiler/analysis frontends should be made with care.

#### M2: Emits documents and ranges

The indexer can emit a validated LSIF index file including a document vertex for each file in the source code and a range vertex for each symbol in the source code. This index should be consumed without error by the latest Sourcegraph instance.

### Compilation unit

The next set of milestones provide definitions, references, and hover text support for symbols within the same _compilation unit_. The granularity of compilation unit depends on the language. For example, files are the compilation unit for Python, packages are the compilation unit for Go.

Data should be provided for symbols contained in the common _80%_ of the language (a rough percentage). Some language features may make the extraction of the correct symbol information difficult, so we do not require 100% of the language to be covered this early in development. Example of possible unsupported features include inner classes in Java, conditional compilation, and the presence of architecture-dependent source files.

#### M3: Emits hover text

The indexer can emit a validated LSIF index file including hover text payloads. This index should be consumed without error by the latest Sourcegraph instance and hover tooltips should appear for these symbols.

#### M4: Emits definitions (within compilation unit)

The indexer can emit a validated LSIF index file including paths from ranges to their _local_ definition (within same compilation unit). This index should be consumed without error by the latest Sourcegraph instance and Go to Definition should work on these symbols.

#### M5: Emits references (within compilation unit)

The indexer can emit a validated LSIF index file including paths from ranges to the set of _local_ references (within same compilation unit). This index should be consumed without error by the latest Sourcegraph instance and single compilation unit Find References should work on these symbols.

At this point, the indexer may be considered **alpha**. It is outputting _some_ useful precise information and Sourcegraph will provide better hover text and _some_ better local code navigation. _Most_ code intel interactions are expected to fall back to search-based code intelligence.

### Input source

The next set of milestones provide definitions and references support for all symbols within the input source code. This should add edges between symbols of different compilation units.

#### M6: Emits definitions (within input source)

The indexer can emit a validated LSIF index file including paths from ranges to their definition, if not defined externally to the input source code. This index should be consumed without error by the latest Sourcegraph instance and Go to Definition should work on these symbols.

#### M7: Emits references (within input source)

The indexer can emit a validated LSIF index file including paths from ranges to the complete set of references contained in the input source code. This index should be consumed without error by the latest Sourcegraph instance and single compilation unit Find References should work on these symbols.

At this point, the indexer may be considered **beta**. It is outputting _more_ useful precise information and Sourcegraph will provide mostly correct code navigation with cross-repository support falling back to search-based code intelligence.

### Cross repository

The next milestone provides support for cross-repository definitions and references.

#### M8: Emits monikers for cross-repository support

The indexer can emit a validated LSIF index file including import monikers for each symbol defined non-locally, and export monikers for each symbol importable by another repository. This index should be consumed without error by the latest Sourcegraph instance and Go to Definition and Find References should work on cross-repository symbols _given that both repositories are indexed at the exact commit imported_.

At this point, the indexer may be generally considered **ready**. Some languages and ecosystems may require some of the additional following milestones to be considered ready due to a bad out-of-the-box developer experience or absence of a critical language features. For example, lsif-java is nearly useless without built-in support for build systems such as gradle, and some customers may reject lsif-clang if it has no support for a language feature introduced in C++14.

### Build tools

The next milestone integrates the indexer with a common build tool or framework for the language and surrounding ecosystem. The priority and applicability of this milestone will vary wildly by languages. For example, lsif-go uses the standard library and the language has a built-in dependency manager; all of our customers use Gradle, making lsif-java effectively unusable without native Gradle support.

#### M9: Common build tool integration

The indexer integrates natively with common mainstream build tools. We should aim to cover _at least_ the majority of build tools used by existing enterprise customers.

The implementation of this milestone will also vary wildly by language. It could be that the indexer runs as a step in a larger tool, or the indexer contains build-specific logic in order to determine the set of source code that needs to be analyzed.

### The long tail

The next milestone represents the never-finished long tail of feature additions. The remaining _20%_ of the language should be supported via incremental updates and releases over time. Support of additional language features should be prioritized by popularity and demand.

#### M10: Support additional language features

We may lack the language expertise or bandwidth to implement certain features on indexers. We will consider investing resources to add additional features when the demand for support is sufficiently high and the implementation is sufficiently difficult.

## Notes on specific indexers

#### lsif-clang

- lsif-clang sometimes outputs nonsensical data when trying to index template usages or definitions.

- C/C++ build system ecosystem is extremely fragmented and difficult to support out of the box. We have decided that integrated build tool support will not be a short-term goal of lsif-clang.

#### lsif-java

- Undergoing a large-scale rewrite not yet on the master branch.
- Native Gradle integration supported in the rewrite.
- Maven support planned after Gradle integration work has concluded.

#### lsif-cpp

- Difficult to install.
- Supports clang only (no gcc builds).
- No hover text support (Sourcegraph will always fall back to search or tooltips).
- Deprecated in favor of lsif-clang.

#### lsif-semanticdb

- Requires generation of a SemanticDB database.
- Support is not currently on the near-term roadmap.

<style type="text/css">
.indexer-implemented-y { text-align: center; background-color: #28a745; }
.indexer-implemented-n { text-align: center; background-color: #dc3545; }
</style>
