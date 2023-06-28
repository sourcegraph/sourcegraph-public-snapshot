package com.sourcegraph.cody.ui;

import static com.sourcegraph.cody.chat.ChatUIConstants.TEXT_MARGIN;

import com.intellij.openapi.editor.colors.EditorColorsManager;
import com.intellij.openapi.editor.colors.EditorColorsScheme;
import com.intellij.ui.BrowserHyperlinkListener;
import com.intellij.ui.ColorUtil;
import com.intellij.util.ui.JBInsets;
import com.intellij.util.ui.SwingHelper;
import com.intellij.util.ui.UIUtil;
import java.awt.Color;
import java.awt.Insets;
import javax.swing.JEditorPane;
import javax.swing.text.html.HTMLEditorKit;
import org.jetbrains.annotations.NotNull;

public class HtmlViewer {
  @NotNull
  public static JEditorPane createHtmlViewer(@NotNull Color backgroundColor) {
    JEditorPane jEditorPane = SwingHelper.createHtmlViewer(true, null, null, null);
    jEditorPane.setEditorKit(new UIUtil.JBWordWrapHtmlEditorKit());
    HTMLEditorKit htmlEditorKit = (HTMLEditorKit) jEditorPane.getEditorKit();
    EditorColorsScheme schemeForCurrentUITheme =
        EditorColorsManager.getInstance().getSchemeForCurrentUITheme();
    String editorFontName = schemeForCurrentUITheme.getEditorFontName();
    int editorFontSize = schemeForCurrentUITheme.getEditorFontSize();
    String fontFamilyAndSize =
        "font-family:'" + editorFontName + "'; font-size:" + editorFontSize + "pt;";
    String backgroundColorCss = "background-color: #" + ColorUtil.toHex(backgroundColor) + ";";
    htmlEditorKit.getStyleSheet().addRule("code { " + backgroundColorCss + fontFamilyAndSize + "}");
    jEditorPane.setFocusable(true);
    jEditorPane.setMargin(
        JBInsets.create(new Insets(TEXT_MARGIN, TEXT_MARGIN, TEXT_MARGIN, TEXT_MARGIN)));
    jEditorPane.addHyperlinkListener(BrowserHyperlinkListener.INSTANCE);
    return jEditorPane;
  }
}
