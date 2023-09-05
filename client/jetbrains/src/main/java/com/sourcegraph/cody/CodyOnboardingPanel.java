package com.sourcegraph.cody;

import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.ui.components.AnActionLink;
import com.intellij.util.ui.JBUI;
import com.sourcegraph.cody.chat.ChatUIConstants;
import com.sourcegraph.cody.chat.ContentWithGradientBorder;
import com.sourcegraph.cody.config.SettingsConfigurable;
import com.sourcegraph.cody.ui.Colors;
import java.awt.*;
import javax.swing.*;
import javax.swing.border.Border;
import org.jetbrains.annotations.NotNull;

public class CodyOnboardingPanel extends JPanel {

  private static final int PADDING = 20;
  // 10 here is the default padding from the styles of the h2 and we want to make the whole padding
  // to be 20, that's why we need the difference between our PADDING and the default padding of the
  // h2
  private static final int ADDITIONAL_PADDING_FOR_HEADER = PADDING - 10;

  public CodyOnboardingPanel(
      @NotNull Project project, @NotNull JEditorPane message, @NotNull JButton button) {
    JPanel panelWithTheMessage =
        new ContentWithGradientBorder(ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH);
    message.setMargin(JBUI.emptyInsets());
    Border paddingInsideThePanel =
        JBUI.Borders.empty(ADDITIONAL_PADDING_FOR_HEADER, PADDING, 0, PADDING);
    panelWithTheMessage.add(message);
    panelWithTheMessage.setBorder(paddingInsideThePanel);
    JPanel buttonPanel = new JPanel(new BorderLayout());
    buttonPanel.add(button, BorderLayout.CENTER);
    buttonPanel.setOpaque(false);
    buttonPanel.setBorder(JBUI.Borders.empty(PADDING, 0));
    panelWithTheMessage.add(buttonPanel);
    this.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
    this.setBorder(JBUI.Borders.empty(PADDING));
    this.add(panelWithTheMessage);
    JPanel goToSettingsPanel = createPanelWithGoToSettingsButton(project);
    this.add(goToSettingsPanel);
  }

  private JPanel createPanelWithGoToSettingsButton(@NotNull Project project) {
    AnActionLink goToSettingsLink =
        new AnActionLink(
            "Sign in with an enterprise account",
            new AnAction() {
              @Override
              public void actionPerformed(@NotNull AnActionEvent e) {
                ShowSettingsUtil.getInstance()
                    .showSettingsDialog(project, SettingsConfigurable.class);
              }
            });
    goToSettingsLink.setForeground(Colors.SECONDARY_LINK_COLOR);
    goToSettingsLink.setAlignmentX(Component.CENTER_ALIGNMENT);
    JPanel panelWithSettingsLink = new JPanel(new BorderLayout());
    panelWithSettingsLink.setBorder(JBUI.Borders.empty(PADDING, 0));
    JPanel linkPanel = new JPanel(new GridBagLayout());
    linkPanel.add(goToSettingsLink);
    panelWithSettingsLink.add(linkPanel, BorderLayout.PAGE_START);
    return panelWithSettingsLink;
  }
}
