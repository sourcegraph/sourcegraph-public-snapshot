package com.sourcegraph.cody.config

import com.intellij.ui.components.JBCheckBox
import com.intellij.util.ui.ColumnInfo
import javax.swing.JTable

class LanguageCheckboxColumn : ColumnInfo<LanguageEntry, Boolean>("") {
    override fun isCellEditable(languageEntry: LanguageEntry): Boolean {
        return true
    }

    override fun getColumnClass(): Class<*> {
        return Boolean::class.java
    }

    override fun valueOf(languageEntry: LanguageEntry): Boolean {
        return !languageEntry.isBlacklisted
    }

    override fun setValue(languageEntry: LanguageEntry, value: Boolean) {
        languageEntry.isBlacklisted = !value
    }

    override fun getWidth(table: JTable): Int {
        return JBCheckBox().preferredSize.width
    }
}

