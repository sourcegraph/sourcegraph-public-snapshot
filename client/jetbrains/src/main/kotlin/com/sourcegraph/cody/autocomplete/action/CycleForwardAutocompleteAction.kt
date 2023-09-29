package com.sourcegraph.cody.autocomplete.action

import com.intellij.openapi.editor.actionSystem.EditorAction

class CycleForwardAutocompleteAction : EditorAction(CycleCodyAutocompleteActionHandler())
