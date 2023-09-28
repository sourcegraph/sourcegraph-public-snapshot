package com.sourcegraph.cody.autocomplete

import com.intellij.openapi.editor.actionSystem.EditorAction

class TriggerAutocompleteAction : EditorAction(TriggerAutocompleteActionHandler())
