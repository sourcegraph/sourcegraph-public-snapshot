package com.sourcegraph.cody.config.ui.lang

import com.intellij.ui.components.JBLabel
import com.intellij.util.ui.JBUI
import java.awt.Component
import javax.swing.JTable
import javax.swing.table.TableCellRenderer

class LanguageNameCellRenderer(
    private val languageTable: AutocompleteLanguageTable,
    val displayName: String? = null
) : TableCellRenderer {
  override fun getTableCellRendererComponent(
      table: JTable?,
      value: Any?,
      isSelected: Boolean,
      hasFocus: Boolean,
      row: Int,
      column: Int
  ): Component {
    val label = JBLabel(displayName ?: value as? String ?: "")
    label.border = JBUI.Borders.empty(0, 8)
    label.isEnabled = languageTable.isEnabled
    return label
  }
}
