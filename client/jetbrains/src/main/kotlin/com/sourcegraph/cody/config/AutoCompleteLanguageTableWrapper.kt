package com.sourcegraph.cody.config

import java.awt.BorderLayout
import java.awt.Dimension
import javax.swing.JPanel

/** Wrapper to be used with JetBrains Kotlin UI DSL */
class AutoCompleteLanguageTableWrapper(private val languageTable: AutocompleteLanguageTable) :
    JPanel(BorderLayout()) {

  init {
    val tableComponent = languageTable.component
    val customHeightDiff = 120
    val customPreferredSize =
        Dimension(
            tableComponent.preferredSize.width,
            tableComponent.preferredSize.height + customHeightDiff)
    tableComponent.preferredSize = customPreferredSize
    add(tableComponent)
  }

  fun getBlacklistedLanguageIds(): List<String> {
    return languageTable.getBlacklistedLanguageIds()
  }

  fun setBlacklistedLanguageIds(blacklistedIds: List<String>) {
    languageTable.setBlacklistedLanguageIds(blacklistedIds)
  }
}
