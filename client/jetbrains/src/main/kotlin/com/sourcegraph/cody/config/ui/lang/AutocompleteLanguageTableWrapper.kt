package com.sourcegraph.cody.config.ui.lang

import java.awt.BorderLayout
import java.awt.Dimension
import javax.swing.JPanel

/** Wrapper to be used with JetBrains Kotlin UI DSL */
class AutocompleteLanguageTableWrapper(private val languageTable: AutocompleteLanguageTable) :
    JPanel(BorderLayout()) {
  private val tableComponent = languageTable.component

  init {
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

  override fun setEnabled(enabled: Boolean) {
    super.setEnabled(enabled)
    tableComponent.isEnabled = enabled
    tableComponent.components.forEach { it.isEnabled = enabled }
    languageTable.isEnabled = enabled
  }
}
