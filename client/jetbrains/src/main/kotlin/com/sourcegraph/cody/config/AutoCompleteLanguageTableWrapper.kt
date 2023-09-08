package com.sourcegraph.cody.config

import java.awt.BorderLayout
import javax.swing.JPanel

/**  Wrapper to be used with JetBrains Kotlin UI DSL */
class AutoCompleteLanguageTableWrapper(private val languageTable: AutocompleteLanguageTable) : JPanel(BorderLayout()) {
    init {
        add(languageTable.component)
    }

    fun getBlacklistedLanguageIds(): List<String> {
        return languageTable.getBlacklistedLanguageIds()
    }

    fun setBlacklistedLanguageIds(blacklistedIds: List<String>) {
        languageTable.setBlacklistedLanguageIds(blacklistedIds)
    }
}
