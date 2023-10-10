package com.sourcegraph.cody.statusbar

import com.intellij.openapi.project.Project
import com.intellij.util.messages.Topic

interface CodyAutocompleteStatusListener {
  fun onCodyAutocompleteStatus(codyAutocompleteStatus: CodyAutocompleteStatus)

  fun onCodyAutocompleteStatusReset(project: Project)

  companion object {
    val TOPIC = Topic.create("cody.autocomplete.status", CodyAutocompleteStatusListener::class.java)
  }
}
