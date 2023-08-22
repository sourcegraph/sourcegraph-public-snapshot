package com.sourcegraph.cody.statusbar

import com.intellij.util.messages.Topic

interface CodyAutocompleteStatusListener {
  fun onCodyAutocompleteStatus(codyAutocompleteStatus: CodyAutocompleteStatus)

  companion object {
    val TOPIC = Topic.create("cody.autocomplete.status", CodyAutocompleteStatusListener::class.java)
  }
}
