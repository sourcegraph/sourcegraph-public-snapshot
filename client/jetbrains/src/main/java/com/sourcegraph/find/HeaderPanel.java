package com.sourcegraph.find;

import com.intellij.icons.AllIcons;
import com.intellij.openapi.actionSystem.*;
import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.util.ui.JBEmptyBorder;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.components.BorderLayoutPanel;
import com.sourcegraph.Icons;
import com.sourcegraph.config.SettingsConfigurable;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;

public class HeaderPanel extends BorderLayoutPanel {
    private final AnAction openAuthenticationSettingsAction;
    private final DefaultActionGroup actionGroup;

    public HeaderPanel(Project project) {
        super();
        setBorder(new JBEmptyBorder(5));

        openAuthenticationSettingsAction = new DumbAwareAction("Set Up Your Sourcegraph Account", "Opens plugin settings page in Settings/Preferences", Icons.Account) {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                ShowSettingsUtil.getInstance().showSettingsDialog(project, SettingsConfigurable.class);
            }
        };
        AnAction openPluginSettingsAction = new DumbAwareAction("Open Plugin Settings", "Opens the plugin settings page in Settings/Preferences", Icons.GearPlain) {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                ShowSettingsUtil.getInstance().showSettingsDialog(project, SettingsConfigurable.class);
            }
        };
        actionGroup = new DefaultActionGroup(openAuthenticationSettingsAction, openPluginSettingsAction);
        ActionToolbar toolbar = ActionManager.getInstance().createActionToolbar("find-on-sourcegraph-popup-toolbar", actionGroup, true);
        toolbar.setMinimumButtonSize(JBUI.size(22, 22));
        toolbar.setTargetComponent(this);
        actionGroup.remove(openAuthenticationSettingsAction);

        JPanel title = new JPanel(new FlowLayout(FlowLayout.LEFT, 0, 0));
        title.add(new JLabel("Find on Sourcegraph", Icons.Logo, SwingConstants.LEFT));

        add(title, BorderLayout.WEST);
        add(toolbar.getComponent(), BorderLayout.EAST);
    }

    public void setAuthenticated(boolean authenticated) {
        if (authenticated) {
            if (actionGroup.containsAction(openAuthenticationSettingsAction)) {
                actionGroup.remove(openAuthenticationSettingsAction);
            }
        } else {
            if (!actionGroup.containsAction(openAuthenticationSettingsAction)) {
                actionGroup.add(openAuthenticationSettingsAction, Constraints.FIRST);
            }
        }

    }
}
