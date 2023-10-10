package com.sourcegraph.cody.vscode

class InlineAutocompleteItem(
    val insertText: String,
    val filterText: String,
    val range: Range,
    val command: Command
)
