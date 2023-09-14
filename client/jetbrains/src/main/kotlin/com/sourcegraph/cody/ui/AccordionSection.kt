package com.sourcegraph.cody.ui

import com.intellij.openapi.ui.VerticalFlowLayout
import java.awt.BorderLayout
import javax.swing.JButton
import javax.swing.JPanel
import javax.swing.SwingConstants

class AccordionSection(title: String) : JPanel() {
  val contentPanel: JPanel
  private val toggleButton: JButton
  private val sectionTitle: String

  init {
    layout = BorderLayout()
    sectionTitle = title
    toggleButton = JButton(createToggleButtonHTML(title, true))
    toggleButton.horizontalAlignment = SwingConstants.LEFT
    toggleButton.isBorderPainted = false
    toggleButton.isFocusPainted = false
    toggleButton.isContentAreaFilled = false
    contentPanel = JPanel()
    toggleButton.addActionListener { _ ->
      if (contentPanel.isVisible) {
        contentPanel.isVisible = false
        toggleButton.text = createToggleButtonHTML(sectionTitle, true)
      } else {
        contentPanel.isVisible = true
        toggleButton.text = createToggleButtonHTML(sectionTitle, false)
      }
    }
    contentPanel.layout = VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false)
    contentPanel.isVisible = false
    add(toggleButton, BorderLayout.NORTH)
    add(contentPanel, BorderLayout.CENTER)
  }

  private fun createToggleButtonHTML(title: String, isCollapsed: Boolean): String {
    val symbol =
        if (isCollapsed) "&#9658;" else "&#9660;" // Unicode entities for right and down arrows
    return ("<html><body style='text-align:left'>" +
        title +
        " <span style='float:right; color:gray'>" +
        symbol +
        "</span></body></html>")
  }
}
