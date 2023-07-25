package com.sourcegraph.cody.ui

import com.intellij.openapi.editor.colors.EditorColorsManager
import com.intellij.ui.BrowserHyperlinkListener
import com.intellij.ui.ColorUtil
import com.intellij.util.ui.JBInsets
import com.intellij.util.ui.SwingHelper
import com.intellij.util.ui.UIUtil.JBWordWrapHtmlEditorKit
import com.sourcegraph.cody.chat.ChatUIConstants
import java.awt.Color
import java.awt.Insets
import javax.swing.JEditorPane
import javax.swing.text.html.HTMLEditorKit

object HtmlViewer {
  @JvmStatic
  fun createHtmlViewer(backgroundColor: Color): JEditorPane {
    val jEditorPane = SwingHelper.createHtmlViewer(true, null, null, null)
    jEditorPane.editorKit = JBWordWrapHtmlEditorKit()
    val htmlEditorKit = jEditorPane.editorKit as HTMLEditorKit
    val fontFamilyAndSize = createFontFamilyAndSizeCssRule()
    val backgroundColorCss = createBackgroundColorCssRule(backgroundColor)
    htmlEditorKit.styleSheet.addRule("code { $backgroundColorCss$fontFamilyAndSize}")
    jEditorPane.isFocusable = true
    jEditorPane.margin =
        JBInsets.create(
            Insets(
                ChatUIConstants.TEXT_MARGIN,
                ChatUIConstants.TEXT_MARGIN,
                ChatUIConstants.TEXT_MARGIN,
                ChatUIConstants.TEXT_MARGIN))
    jEditorPane.addHyperlinkListener(BrowserHyperlinkListener.INSTANCE)
    return jEditorPane
  }

  private fun createBackgroundColorCssRule(backgroundColor: Color) =
      "background-color: #" + ColorUtil.toHex(backgroundColor) + ";"

  private fun createFontFamilyAndSizeCssRule(): String {
    val schemeForCurrentUITheme = EditorColorsManager.getInstance().schemeForCurrentUITheme
    val editorFontName = schemeForCurrentUITheme.editorFontName
    val editorFontSize = schemeForCurrentUITheme.editorFontSize
    return "font-family:'" + editorFontName + "'; font-size:" + editorFontSize + "pt;"
  }
}
