package com.sourcegraph.cody.config

import com.intellij.ui.components.JBLabel
import com.intellij.util.ui.ColumnInfo
import java.util.*
import javax.swing.table.TableCellRenderer

class LanguageEntryColumn(private val languageTable: AutocompleteLanguageTable) :
    ColumnInfo<LanguageEntry, String>("Enabled Languages") {
  override fun getCustomizedRenderer(
      o: LanguageEntry?,
      renderer: TableCellRenderer?
  ): TableCellRenderer {
    // the language entry shouldn't ever in fact be a null here
    val label = JBLabel(o?.language?.displayName ?: "Unknown")
    label.isEnabled = languageTable.isEnabled
    return TableCellRenderer { _, _, _, _, _, _ -> label }
  }

  override fun valueOf(languageEntry: LanguageEntry): String {
    return languageEntry.language.displayName
  }

  override fun getComparator(): Comparator<LanguageEntry>? {
    return Comparator.comparing { le: LanguageEntry -> le.language.displayName.lowercase() }
  }
}
