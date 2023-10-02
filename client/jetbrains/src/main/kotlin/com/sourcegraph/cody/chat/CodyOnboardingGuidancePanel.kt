package com.sourcegraph.cody.chat

import com.intellij.ide.ui.laf.darcula.ui.DarculaButtonUI
import com.intellij.ui.ColorUtil
import com.intellij.ui.JBColor
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBScrollPane
import com.intellij.util.ui.JBUI
import com.intellij.util.ui.UIUtil
import com.sourcegraph.cody.Icons
import com.sourcegraph.cody.chat.UIComponents.createMainButton
import com.sourcegraph.cody.ui.HtmlViewer.createHtmlViewer
import java.awt.BorderLayout
import java.awt.Dimension
import java.awt.event.ActionListener
import javax.swing.BorderFactory
import javax.swing.BoxLayout
import javax.swing.Icon
import javax.swing.JButton
import javax.swing.JComponent
import javax.swing.JEditorPane
import javax.swing.JPanel
import javax.swing.text.html.HTMLEditorKit
import org.apache.commons.lang3.StringUtils

class CodyOnboardingGuidancePanel(val originalDisplayName: String?) : JPanel() {

  private val userDisplayName: String? = originalDisplayName?.let { truncateDisplayName(it) }

  private val mainButton: JButton = createMainButton("Get started")

  init {
    val contentPanel = JPanel()
    contentPanel.layout = BoxLayout(contentPanel, BoxLayout.Y_AXIS)
    contentPanel.add(createChatWithCodyPanel())
    contentPanel.add(createAutocompletePanel())
    contentPanel.add(createExploreCommandsPanel())

    val scrollPanel =
        JBScrollPane(
            contentPanel,
            JBScrollPane.VERTICAL_SCROLLBAR_AS_NEEDED,
            JBScrollPane.HORIZONTAL_SCROLLBAR_NEVER)
    scrollPanel.preventStretching()
    scrollPanel.setBorder(BorderFactory.createEmptyBorder())

    val buttonPanel = createGetStartedButton()
    this.border = JBUI.Borders.empty(PADDING)
    this.layout = BoxLayout(this, BoxLayout.Y_AXIS)
    this.add(createIntroductionMessage())
    this.add(scrollPanel)
    this.add(buttonPanel)
  }

  private fun createGreetings(): String {
    if (!userDisplayName.isNullOrEmpty()) {
      return "Hi, $userDisplayName"
    }
    return "Hi"
  }

  private fun createIntroductionMessage(): JEditorPane {
    val introductionMessage = createHtmlViewer(UIUtil.getPanelBackground())
    val introductionMessageEditorKit = introductionMessage.editorKit as HTMLEditorKit
    introductionMessageEditorKit.styleSheet.addRule(paragraphColorStyle)
    introductionMessage.text = buildString {
      append("<html><body><h2>${createGreetings()}</h2>")
      append("<p>Let's start by getting you familiar with all the possibilities Cody provides:</p>")
      append("</body></html>")
    }
    introductionMessage.setMargin(JBUI.emptyInsets())
    introductionMessage.preventStretching()
    return introductionMessage
  }

  private fun createGetStartedButton(): JPanel {
    val buttonPanel = JPanel(BorderLayout())
    mainButton.putClientProperty(DarculaButtonUI.DEFAULT_STYLE_KEY, true)
    buttonPanel.add(mainButton, BorderLayout.NORTH)
    buttonPanel.border = BorderFactory.createEmptyBorder(PADDING, 0, 0, 0)
    return buttonPanel
  }

  private fun createSectionWithTextAndImage(sectionText: String, sectionImage: Icon?): JPanel {
    val sectionPanel = sectionPanel()
    val infoSectionEditorPane = createInfoSection()
    infoSectionEditorPane.text = sectionText
    infoSectionEditorPane.setMargin(JBUI.insets(PADDING))
    sectionPanel.add(infoSectionEditorPane, BorderLayout.NORTH)
    val imagePanel = JPanel(BorderLayout())
    imagePanel.border = BorderFactory.createEmptyBorder(0, PADDING, PADDING, PADDING)
    imagePanel.add(JBLabel(sectionImage), BorderLayout.SOUTH)
    sectionPanel.add(imagePanel)
    return sectionPanel
  }

  private fun createChatWithCodyPanel() =
      createSectionWithTextAndImage(
          buildString {
            append("<html><body><h3>Chat with Cody</h3>")
            append(
                "<p>Use this sidebar to engage with Cody. Get <b>answers</b> and suggestions about the code you are working on.</p>")
            append("</body></html>")
          },
          Icons.Onboarding.Chat)

  private fun createAutocompletePanel() =
      createSectionWithTextAndImage(
          buildString {
            append("<html><body><h3>Autocompletions</h3>")
            append(
                "<p>Start typing code to get <b>autocompletions</b> base on the surrounding context (press Tab to accept them):</p>")
            append("</body></html>")
          },
          Icons.Onboarding.Autocomplete)

  private fun createExploreCommandsPanel() =
      createSectionWithTextAndImage(
          "<html><body><h3>Explore the Commands</h3>" +
              "<p>Use <b>commands</b> to execute useful tasks on your code, like generating unit tests, docstrings and more</p>" +
              "</body></html>",
          Icons.Onboarding.Commands)

  private fun JComponent.preventStretching() {
    maximumSize = Dimension(Int.MAX_VALUE, getPreferredSize().height)
  }

  private fun sectionPanel(): JPanel {
    val panel = JPanel()
    panel.layout = BorderLayout()
    panel.border =
        BorderFactory.createCompoundBorder(
            BorderFactory.createEmptyBorder(PADDING, 0, 0, 0),
            BorderFactory.createLineBorder(borderColor, 1, true))
    return panel
  }

  private fun createInfoSection(): JEditorPane {
    val sectionInfo = createHtmlViewer(UIUtil.getPanelBackground())
    val sectionInfoHtmlEditorKit = sectionInfo.editorKit as HTMLEditorKit
    sectionInfoHtmlEditorKit.styleSheet.addRule(paragraphColorStyle)
    sectionInfoHtmlEditorKit.styleSheet.addRule("""h3 { margin-top: 0;}""")
    return sectionInfo
  }

  private fun truncateDisplayName(displayName: String): String {
    if (displayName.length > 32) {
      return StringUtils.truncate(displayName, 32) + "..."
    }
    return displayName
  }

  fun addMainButtonActionListener(actionListener: ActionListener) {
    mainButton.addActionListener(actionListener)
  }

  companion object {
    private const val PADDING = 20

    private val borderColor =
        JBColor(
            ColorUtil.darker(UIUtil.getPanelBackground(), 1),
            ColorUtil.brighter(UIUtil.getPanelBackground(), 3))
    private val paragraphColor =
        JBColor(
            ColorUtil.brighter(UIUtil.getLabelForeground(), 2),
            ColorUtil.darker(UIUtil.getLabelForeground(), 2))
    private val paragraphColorStyle = """p { color: ${ColorUtil.toHtmlColor(paragraphColor)}; }"""
  }
}
