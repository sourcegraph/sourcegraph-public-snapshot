package com.sourcegraph.find;

import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.Presentation;
import com.intellij.openapi.actionSystem.impl.ActionButton;
import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.util.IconUtil;
import com.intellij.util.ui.JBDimension;
import com.intellij.util.ui.JBEmptyBorder;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.components.BorderLayoutPanel;
import com.sourcegraph.Icons;
import com.sourcegraph.config.SettingsConfigurable;
import java.awt.*;
import javax.swing.*;
import org.jetbrains.annotations.NotNull;

public class HeaderPanel extends BorderLayoutPanel {
  public HeaderPanel(Project project) {
    super();
    setBorder(new JBEmptyBorder(5, 5, 2, 5));

    JPanel title = new JPanel(new FlowLayout(FlowLayout.LEFT, 0, 0));
    title.setBorder(new JBEmptyBorder(2, 0, 0, 0));
    title.add(new JLabel("Find with Sourcegraph", Icons.SourcegraphLogo, SwingConstants.LEFT));

    JPanel buttons = new JPanel(new FlowLayout(FlowLayout.RIGHT, 0, 0));
    buttons.add(createSettingsButton(project));

    add(title, BorderLayout.WEST);
    add(buttons, BorderLayout.EAST);
  }

  @NotNull
  private ActionButton createSettingsButton(@NotNull Project project) {
    JBDimension actionButtonSize = JBUI.size(22, 22);

    AnAction action =
        new DumbAwareAction() {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            ShowSettingsUtil.getInstance().showSettingsDialog(project, SettingsConfigurable.class);
          }
        };
    Presentation presentation = new Presentation("Open Plugin Settings");

    ActionButton button =
        new ActionButton(
            action, presentation, "Find with Sourcegraph popup header", actionButtonSize);

    Icon scaledIcon = IconUtil.scale(Icons.GearPlain, button, 13f / 12f);
    presentation.setIcon(scaledIcon);

    return button;
  }
}
