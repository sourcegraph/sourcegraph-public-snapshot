package com.sourcegraph.cody;

import static com.sourcegraph.cody.chat.ChatUIConstants.TEXT_MARGIN;

import com.intellij.ide.ui.laf.darcula.ui.DarculaButtonUI;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.VerticalFlowLayout;
import com.intellij.ui.components.AnActionLink;
import com.intellij.util.ui.JBUI;
import com.sourcegraph.cody.chat.ChatUIConstants;
import com.sourcegraph.cody.chat.ContentWithGradientBorder;
import com.sourcegraph.cody.ui.Colors;
import com.sourcegraph.config.SettingsConfigurable;
import java.awt.*;
import javax.swing.*;
import javax.swing.border.Border;
import javax.swing.plaf.ButtonUI;
import org.jetbrains.annotations.NotNull;

public class CodyOnboardingPanel extends JPanel {

  public CodyOnboardingPanel(
      @NotNull Project project, @NotNull JEditorPane message, @NotNull JButton button) {
    JPanel panelWithTheMessage =
        new ContentWithGradientBorder(ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH);
    panelWithTheMessage.add(message);
    Border margin = JBUI.Borders.empty(TEXT_MARGIN);
    message.setBorder(margin);
    JPanel buttonPanel = new JPanel(new BorderLayout());
    buttonPanel.add(button, BorderLayout.CENTER);
    buttonPanel.setOpaque(false);
    buttonPanel.setBorder(margin);
    panelWithTheMessage.add(buttonPanel);
    JPanel blankPanel = new JPanel();
    blankPanel.setBorder(margin);
    blankPanel.setOpaque(false);
    panelWithTheMessage.add(blankPanel);
    this.setLayout(new VerticalFlowLayout(VerticalFlowLayout.TOP, 0, 0, true, false));
    this.setBorder(margin);
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
    ButtonUI buttonUI = (ButtonUI) DarculaButtonUI.createUI(goToSettingsLink);
    goToSettingsLink.setUI(buttonUI);
    goToSettingsLink.updateUI();
    JPanel panelWithSettingsLink = new JPanel(new BorderLayout());
    panelWithSettingsLink.setBorder(JBUI.Borders.empty(TEXT_MARGIN, 0));
    JPanel linkPanel = new JPanel(new GridBagLayout());
    linkPanel.add(goToSettingsLink);
    panelWithSettingsLink.add(linkPanel, BorderLayout.PAGE_START);
    return panelWithSettingsLink;
  }
}
