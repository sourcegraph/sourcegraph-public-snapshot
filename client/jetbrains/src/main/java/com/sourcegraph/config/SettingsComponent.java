package com.sourcegraph.config;

import com.intellij.ui.IdeBorderFactory;
import com.intellij.ui.components.JBCheckBox;
import com.intellij.ui.components.JBLabel;
import com.intellij.ui.components.JBTextField;
import com.intellij.util.ui.FormBuilder;
import com.intellij.util.ui.JBUI;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;

/**
 * Supports creating and managing a {@link JPanel} for the Settings Dialog.
 */
public class SettingsComponent {

    private final JPanel panel;
    private JBTextField sourcegraphUrlTextField;
    private JBTextField accessTokenTextField;
    private JBTextField defaultBranchNameTextField;
    private JBTextField remoteUrlReplacementsTextField;
    private JBCheckBox globbingCheckBox;
    private JBCheckBox isUrlNotificationDismissedCheckBox;

    public SettingsComponent() {
        JPanel userAuthenticationPanel = createAuthenticationPanel();
        JPanel navigationSettingsPanel = createNavigationSettingsPanel();

        panel = FormBuilder.createFormBuilder()
            .addComponent(userAuthenticationPanel)
            .addComponent(navigationSettingsPanel)
            .addComponentFillVertically(new JPanel(), 0)
            .getPanel();
    }

    public JPanel getPanel() {
        return panel;
    }

    public JComponent getPreferredFocusedComponent() {
        return sourcegraphUrlTextField;
    }

    @NotNull
    public String getSourcegraphUrl() {
        return sourcegraphUrlTextField.getText();
    }

    public void setSourcegraphUrl(@NotNull String value) {
        sourcegraphUrlTextField.setText(value);
    }

    @NotNull
    public String getAccessToken() {
        return accessTokenTextField.getText();
    }

    public void setAccessToken(@NotNull String value) {
        accessTokenTextField.setText(value);
    }

    @NotNull
    public String getDefaultBranchName() {
        return defaultBranchNameTextField.getText();
    }

    public void setDefaultBranchName(@NotNull String value) {
        defaultBranchNameTextField.setText(value);
    }

    @NotNull
    public String getRemoteUrlReplacements() {
        return remoteUrlReplacementsTextField.getText();
    }

    public void setRemoteUrlReplacements(@NotNull String value) {
        remoteUrlReplacementsTextField.setText(value);
    }

    public boolean isGlobbingEnabled() {
        return globbingCheckBox.isSelected();
    }

    public void setGlobbingEnabled(boolean value) {
        globbingCheckBox.setSelected(value);
    }

    public boolean isUrlNotificationDismissed() {
        return isUrlNotificationDismissedCheckBox.isSelected();
    }

    public void setUrlNotificationDismissedEnabled(boolean value) {
        isUrlNotificationDismissedCheckBox.setSelected(value);
    }

    @NotNull
    private JPanel createAuthenticationPanel() {
        JBLabel sourcegraphUrlLabel = new JBLabel("Sourcegraph URL:");
        sourcegraphUrlTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        sourcegraphUrlTextField.getEmptyText().setText("https://sourcegraph.example.com");
        sourcegraphUrlTextField.setToolTipText("The default is \"https://sourcegraph.com\".");

        JBLabel accessTokenLabel = new JBLabel("Access token:");
        accessTokenTextField = new JBTextField();
        accessTokenTextField.getEmptyText().setText("Paste your access token here");
        accessTokenTextField.setToolTipText("");

        JPanel userAuthenticationPanel = FormBuilder.createFormBuilder()
            .addLabeledComponent(sourcegraphUrlLabel, sourcegraphUrlTextField)
            .addTooltip("If your company has your own Sourcegraph instance, set its URL here")
            .addLabeledComponent(accessTokenLabel, accessTokenTextField)
            .addTooltip("Go to https://sourcegraph.example.com/settings | \"Access tokens\" to create a token")
            .getPanel();
        userAuthenticationPanel.setBorder(IdeBorderFactory.createTitledBorder("User Authentication", true, JBUI.insetsTop(8)));
        return userAuthenticationPanel;
    }

    @NotNull
    private JPanel createNavigationSettingsPanel() {
        JBLabel defaultBranchNameLabel = new JBLabel("Default branch name:");
        defaultBranchNameTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        defaultBranchNameTextField.getEmptyText().setText("main");
        defaultBranchNameTextField.setToolTipText("Usually \"main\" or \"master\", but can be any name");

        JBLabel remoteUrlReplacementsLabel = new JBLabel("Remote URL replacements:");
        remoteUrlReplacementsTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        remoteUrlReplacementsTextField.getEmptyText().setText("search1, replacement1, search2, replacement2, ...");
        //remoteUrlReplacementsTextField.setToolTipText();

        globbingCheckBox = new JBCheckBox("Enable globbing");
        isUrlNotificationDismissedCheckBox = new JBCheckBox("Never show the \"No Sourcegraph URL set\" notification for this project");

        JPanel navigationSettingsPanel = FormBuilder.createFormBuilder()
            .addLabeledComponent(defaultBranchNameLabel, defaultBranchNameTextField)
            .addTooltip("The branch to use if the current branch is not yet pushed")
            .addLabeledComponent(remoteUrlReplacementsLabel, remoteUrlReplacementsTextField)
            .addTooltip("You can replace specified strings in your repo's remote URL.")
            .addTooltip("The format is \"search1, replacement1, search2, replacement2, ...\".")
            .addTooltip("Use any number of search/replacement pairs. Pairs will be replaced going from left to right.")
            .addTooltip("The whitespace before and after the commas doesn't matter.")
            .addComponent(globbingCheckBox)
            .addComponent(isUrlNotificationDismissedCheckBox, 10)
            .getPanel();
        navigationSettingsPanel.setBorder(IdeBorderFactory.createTitledBorder("Navigation Settings", true, JBUI.insetsTop(8)));
        return navigationSettingsPanel;
    }
}
