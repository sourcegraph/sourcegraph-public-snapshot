package com.sourcegraph.cody.config.ui.lang

import com.intellij.util.ui.ColumnInfo
import java.util.*
import javax.swing.table.TableCellRenderer

class LanguageEntryColumn(private val languageTable: AutocompleteLanguageTable) :
    ColumnInfo<LanguageEntry, String>("Enabled Languages") {
  override fun getCustomizedRenderer(
      o: LanguageEntry?,
      renderer: TableCellRenderer?
  ): TableCellRenderer = LanguageNameCellRenderer(languageTable, o?.language?.displayName)

  override fun valueOf(languageEntry: LanguageEntry): String {
    return languageEntry.language.displayName
  }

  override fun getComparator(): Comparator<LanguageEntry>? {
    return Comparator.comparing { le: LanguageEntry -> le.language.displayName.lowercase() }
  }
}
