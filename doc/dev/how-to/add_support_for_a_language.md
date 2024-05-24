# How to add support for a language

The following guides detail the steps necessary to add or upgrade support for a particular programming language.

## Symbols support

To support symbol search and the symbols sidebar:

1. Add or update the target language's configuration in [sourcegraph/go-ctags](https://github.com/sourcegraph/go-ctags)/[.ctags.d](https://github.com/sourcegraph/go-ctags/tree/main/ctagsdotd). The [universal-ctags/ctags](https://github.com/universal-ctags/ctags) project bundles configuration for many languages, but additional or override configuration may be necessary to support missing or correct incorrectly parsed language features. Examples:
    - [scala](https://github.com/sourcegraph/go-ctags/blob/main/ctagsdotd/scala.ctags) (new language)
    - [clojure](https://github.com/sourcegraph/go-ctags/blob/main/ctagsdotd/clojure.ctags) (additional patterns)
    - [css](https://github.com/sourcegraph/go-ctags/blob/main/ctagsdotd/css.ctags) (additional file extensions)
1. Update the [sourcegraph/go-ctags](https://github.com/sourcegraph/go-ctags) dependency in [sourcegraph/zoekt](https://github.com/sourcegraph/zoekt).
1. Update the [sourcegraph/go-ctags](https://github.com/sourcegraph/go-ctags) dependency in [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph).
1. Run `./dev/zoekt/update` in [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph) to pull in the new zoekt version.

#### Development

You can validate your changes to the `.ctag.d` definitions by observing the symbol search and symbol sidebar behaviors in a local Sourcegraph. Link your Sourcegraph instance and your local clone of go-ctags by adding the following to the `go.mod` file in [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph).

```
replace github.com/sourcegraph/go-ctags => /path/to/my/clone/of/sourcegraph/go-ctags
```

Remember to run the code-generation step in the go-ctags repository and restart your local instance after each change to a definition file.

**Note**: Because we have not yet updated Zoekt, you will need to ensure that your symbol search and symbol sidebar operations are not on an indexed branch. These paths must hit the _unindexed_ symbols paths in order to hit the code path with updated ctags definitions.

## Code navigation support

To support precise code navigation, [write an SCIP indexer](../../code_navigation/explanations/writing_an_indexer.md). To support search-based code navigation, ensure the language is registered in the code navigation APIs:

0. _Code navigation support are powered by symbol search. If the target language is not supported by symbols, stop and follow the guide above first._
1. Add (or update) the target language's configuration in [languages.ts](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/client/shared/src/codeintel/legacy-extensions/language-specs/languages.ts#L360). See the definition of [LanguageSpec](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/client/shared/src/codeintel/legacy-extensions/language-specs/language-spec.ts#L7) for an available set of fields. The likely differences will be the characters that make up the identifier, the comment delimiters, and the set of file extensions to search within for definitions and references.
1. Correlate the language's file extensions and the new Sourcegraph extension by adding entries to the switch in [getModeFromExtension](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/client/shared/src/languages.ts?L44#L40:10). This enables the providers from the code navigation APIs to be registered when a text document with the correlated language is opened. The value returned from this function and the `languageId` from the language's configuration should match exactly.

## Syntax highlighting support

### Customizing Syntax Highlight Language

The following settings apply only to the site settings. They are global configuration options for your Sourcegraph instance.

If you have a custom language that is derived from an existing language, it is possible to configure Sourcegraph to highlight that language as another.

For example:

```json
{
  "syntaxHighlighting": {
   "languages": {
     "extensions": {
        "strato": "scala"
      },
      "patterns": []
    }
  }
}
```

If you have custom file extensions that map to an existing language, it is possible to configure Sourcegraph to highlight those files as an existing language.

For example:

```json
{
  "syntaxHighlighting": {
   "languages": {
     "extensions": {
        "module": "php",
        "inc": "php"
      },
      "patterns": []
    }
  }
}
```

NOTE: In both cases, the `.` is dropped from the file extension.

Additionally, for more complex matching, it possible to pass regexes that will be evaluated (in order listed in the configuration) and if a match is found, will override the syntax highlight language for that file.

For example:

```json
{
  "syntaxHighlighting": {
    "languages": {
      "extensions": {},
      "patterns": [
        {
          "language": "bash",
          "pattern": "bash.rc"
        },
        {
          "language": "bash",
          "pattern": ".bashprofile"
        }
      ]
    }
  }
}
```


### Adding New Syntax Highlighting


To add syntax highlighting for a language follow the steps below. You can also refer to this [PR](https://github.com/sourcegraph/sourcegraph/pull/61478) that added highlighting support for the Pkl language.


**Under [docker-images/syntax-highlighter/crates](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/docker-images/syntax-highlighter/crates):**

1. In  [tree-sitter-all-languages](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/docker-images/syntax-highlighter/crates/tree-sitter-all-languages/README.md) add a [Tree-sitter](https://tree-sitter.github.io/tree-sitter/) grammar dependency for the language in the crate (e.g. `cargo add --git https://github.com/example/repo.git --rev abcdef1234` ).

1. In [tree-sitter-all-languages/lib.rs](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/docker-images/syntax-highlighter/crates/tree-sitter-all-languages/src/lib.rs) add new entries in the `ParserId` enum and associated references. 

1. In [syntax-analysis/queries](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/docker-images/syntax-highlighter/crates/syntax-analysis/queries) add a folder for the language with three files in it (`highlights.scm`, `locals.scm`, and `injections.scm`). These files can be empty besides `highlights.scm`, which must contain tree-sitter queries for identifying and labeling all tokens in the language requiring highlighting. For more information on writing tree-sitter queries, see the [tree-sitter docs](https://tree-sitter.github.io/tree-sitter/syntax-highlighting#highlights) or refer to the other `.scm` files in the folder.

1. In [syntax-analysis/src/highlighting/snapshots/files](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/docker-images/syntax-highlighter/crates/syntax-analysis/src/highlighting/snapshots/files) add an example file for the language which demonstrates all key language elements needing highlighting. Some tips for writing queries and adding test cases:

  - Try to be exhaustive: Exercise as many syntax features as possible. Generally, this involves going through the language's reference documentation, identifying the various syntax features that are available, and making sure there are queries and code examples for each.
  - Avoid redundancy: Generally, every line should exercise a syntax feature that hasn't been exercised earlier in the file. More lines increases code review burden, and when the snapshots change, redundancy creates larger diffs.
  - Explicitly state known gaps: If a syntax feature is not supported by the grammar, file an issue for that on the grammar's repository and link that issue from a comment in the example file next to that syntax feature.

1.  In [syntax-analysis/src/highlighting](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/docker-images/syntax-highlighter/crates/syntax-analysis) run cargo tests `cargo test` to execute the tests and `cargo insta review` to update and regenerate new snapshots. Please review the snapshot generated and ensure the file has proper highlighting annotations ([example](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/docker-images/syntax-highlighter/crates/syntax-analysis/src/highlighting/snapshots/syntax_analysis__highlighting__tree_sitter__test__python.py.snap)). 
 

**Under [internal/gosyntect/languages.go](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/internal/gosyntect/languages.go):**
1. Add the new language to the list of tree-sitter supported file types. 


**(Optional) If the language is not yet supported in [go-enry](https://github.com/go-enry/go-enry)**

Under [lib/codeintel/languages/extensions.go](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/lib/codeintel/languages/extensions.go?L67):
1. Add a mapping of the language's extension to name in the `unsupportedByEnryExtensionsMap` and update associated unit test

 
