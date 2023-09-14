package com.sourcegraph.cody.config

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
  init {
    setValues(LanguageEntry.getRegisteredLanguageEntries())
  }

  /** Use this rather than the component directly when working with JetBrains Kotlin UI DSL */
  val wrapperComponent: AutoCompleteLanguageTableWrapper = AutoCompleteLanguageTableWrapper(this)

  override fun createListModel(): ListTableModel<LanguageEntry> {
    val model =
        ListTableModel(
            arrayOf<ColumnInfo<*, *>>(LanguageCheckboxColumn(), LanguageEntryColumn()),
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
