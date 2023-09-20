package com.sourcegraph.cody.chat

import com.intellij.ide.ui.laf.darcula.ui.DarculaButtonUI
import com.intellij.openapi.ui.VerticalFlowLayout
import com.intellij.ui.ColorUtil
import com.intellij.ui.SeparatorComponent
import com.intellij.ui.components.AnActionLink
import com.intellij.util.ui.JBUI
import com.intellij.util.ui.UIUtil
import com.sourcegraph.cody.Icons
import com.sourcegraph.cody.auth.ui.SignInWithSourcegraphAction
import com.sourcegraph.cody.ui.HtmlViewer.createHtmlViewer
import com.sourcegraph.cody.ui.UnderlinedActionLink
import java.awt.BorderLayout
import java.awt.FlowLayout
import java.awt.GridBagLayout
import java.awt.event.ActionListener
import javax.swing.BoxLayout
import javax.swing.JLabel
import javax.swing.JPanel
import javax.swing.border.Border

class SignInWithSourcegraphPanel : JPanel() {

  private val mainButton = UIComponents.createMainButton("Sign in for free with Sourcegraph.com")

  init {
    val jEditorPane = createHtmlViewer(UIUtil.getPanelBackground())
    jEditorPane.text =
        ("<html><body><h2>Welcome to Cody</h2>" +
            "<p>Understand and write code faster with an AI assistant</p>" +
            "</body></html>")
    val signInWithSourcegraphButton = mainButton
    signInWithSourcegraphButton.putClientProperty(DarculaButtonUI.DEFAULT_STYLE_KEY, true)

    val panelWithTheMessage = JPanel()
    panelWithTheMessage.setLayout(BoxLayout(panelWithTheMessage, BoxLayout.Y_AXIS))
    jEditorPane.setMargin(JBUI.emptyInsets())
    val paddingInsideThePanel: Border =
        JBUI.Borders.empty(ADDITIONAL_PADDING_FOR_HEADER, PADDING, 0, PADDING)
    val hiImCodyLabel = JLabel(Icons.HiImCody)
    val hiImCodyPanel = JPanel(FlowLayout(FlowLayout.LEFT, 5, 0))
    hiImCodyPanel.add(hiImCodyLabel)
    panelWithTheMessage.add(hiImCodyPanel)
    panelWithTheMessage.add(jEditorPane)
    panelWithTheMessage.setBorder(paddingInsideThePanel)
    val separatorPanel = JPanel(BorderLayout())
    separatorPanel.setBorder(JBUI.Borders.empty(PADDING, 0))
    val separatorComponent =
        SeparatorComponent(
            3, ColorUtil.brighter(UIUtil.getPanelBackground(), 3), UIUtil.getPanelBackground())
    separatorPanel.add(separatorComponent)
    panelWithTheMessage.add(separatorPanel)
    val buttonPanel = JPanel(BorderLayout())
    buttonPanel.add(signInWithSourcegraphButton, BorderLayout.CENTER)
    buttonPanel.setOpaque(false)
    panelWithTheMessage.add(buttonPanel)
    setLayout(VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false))
    setBorder(JBUI.Borders.empty(PADDING))
    this.add(panelWithTheMessage)
    this.add(createPanelWithSignInWithAnEnterpriseInstance())
  }

  fun addMainButtonActionListener(actionListener: ActionListener) {
    mainButton.addActionListener(actionListener)
  }

  private fun createPanelWithSignInWithAnEnterpriseInstance(): JPanel {
    val signInWithAnEnterpriseInstance: AnActionLink =
        UnderlinedActionLink("Sign in with an Enterprise Instance", SignInWithSourcegraphAction(""))
    signInWithAnEnterpriseInstance.setAlignmentX(CENTER_ALIGNMENT)
    val panelWithSettingsLink = JPanel(BorderLayout())
    panelWithSettingsLink.setBorder(JBUI.Borders.empty(PADDING, 0))
    val linkPanel = JPanel(GridBagLayout())
    linkPanel.add(signInWithAnEnterpriseInstance)
    panelWithSettingsLink.add(linkPanel, BorderLayout.PAGE_START)
    return panelWithSettingsLink
  }

  companion object {
    private const val PADDING = 20

    // 10 here is the default padding from the styles of the h2 and we want to make the whole
    // padding
    // to be 20, that's why we need the difference between our PADDING and the default padding of
    // the
    // h2
    private const val ADDITIONAL_PADDING_FOR_HEADER = PADDING - 10
  }
}
