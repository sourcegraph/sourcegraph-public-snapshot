package com.sourcegraph.cody.config.ui.lang

import com.intellij.execution.util.ListTableWithButtons
import com.intellij.ui.AnActionButtonRunnable
import com.intellij.util.ui.ColumnInfo
import com.intellij.util.ui.ListTableModel
import java.util.stream.Collectors
import javax.swing.SortOrder

/**
 * This table shows languages in whitelist format specifically for the UI. We actually blacklist
 * language entries under the hood.
 */
class AutocompleteLanguageTable : ListTableWithButtons<LanguageEntry>() {
  var isEnabled: Boolean = true

  /** Use this rather than the component directly when working with JetBrains Kotlin UI DSL */
  val wrapperComponent: AutocompleteLanguageTableWrapper = AutocompleteLanguageTableWrapper(this)

  init {
    setValues(LanguageEntry.getRegisteredLanguageEntries())
    tableView.columnModel.columns.asIterator().forEach {
      it.headerRenderer = LanguageNameCellRenderer(this)
    }
  }

  override fun createListModel(): ListTableModel<LanguageEntry> {
    val model =
        ListTableModel(
            arrayOf<ColumnInfo<*, *>>(LanguageCheckboxColumn(this), LanguageEntryColumn(this)),
            listOf<LanguageEntry>(),
            1,
            SortOrder.ASCENDING)
    model.isSortable = true
    return model
  }

  override fun createElement(): LanguageEntry {
    return LanguageEntry.ANY // placeholder
  }

  override fun isEmpty(element: LanguageEntry): Boolean {
    return false
  }

  override fun cloneElement(languageEntry: LanguageEntry): LanguageEntry {
    return LanguageEntry(languageEntry.language, languageEntry.isBlacklisted)
  }

  override fun canDeleteElement(selection: LanguageEntry): Boolean {
    return false
  }

  override fun createAddAction(): AnActionButtonRunnable? {
    return null
  }

  override fun createRemoveAction(): AnActionButtonRunnable? {
    return null
  }

  fun getBlacklistedLanguageIds(): List<String> {
    return elements
        .stream()
        .filter { le -> le.isBlacklisted }
        .map { le -> le.language.id }
        .collect(Collectors.toList())
  }

  fun setBlacklistedLanguageIds(blacklistedIds: List<String>) {
    elements.stream().forEach { le ->
      if (blacklistedIds.contains(le.language.id)) le.isBlacklisted = true
    }
  }
}
