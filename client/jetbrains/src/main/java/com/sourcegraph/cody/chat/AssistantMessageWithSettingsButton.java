package com.sourcegraph.cody.chat;

import static com.sourcegraph.cody.chat.ChatUIConstants.TEXT_MARGIN;

import com.intellij.ui.BrowserHyperlinkListener;
import com.intellij.util.ui.JBInsets;
import com.intellij.util.ui.SwingHelper;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.config.GoToPluginSettingsButtonFactory;
import java.awt.*;
import javax.swing.*;
import org.jetbrains.annotations.NotNull;

public class AssistantMessageWithSettingsButton extends PanelWithGradientBorder {
  public AssistantMessageWithSettingsButton(@NotNull String htmlTextContent) {
    super(ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH, Speaker.ASSISTANT);
    this.setLayout(new BorderLayout());
    JEditorPane jEditorPane = SwingHelper.createHtmlViewer(true, null, null, null);
    jEditorPane.addHyperlinkListener(BrowserHyperlinkListener.INSTANCE);
    jEditorPane.setFocusable(true);
    jEditorPane.setMargin(
        JBInsets.create(new Insets(TEXT_MARGIN, TEXT_MARGIN, TEXT_MARGIN, TEXT_MARGIN)));
    SwingHelper.setHtml(jEditorPane, htmlTextContent, UIUtil.getActiveTextColor());
    this.add(jEditorPane, BorderLayout.CENTER);
    this.add(GoToPluginSettingsButtonFactory.createGoToPluginSettingsButton(), BorderLayout.EAST);
  }
}
