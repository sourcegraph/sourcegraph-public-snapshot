package com.sourcegraph.config;

import com.intellij.ui.components.JBCheckBox;
import com.intellij.ui.components.JBLabel;
import com.intellij.ui.components.JBTextField;
import com.intellij.util.ui.FormBuilder;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;

/**
 * Supports creating and managing a {@link JPanel} for the Settings Dialog.
 */
public class SettingsComponent {

    private final JPanel panel;
    private final JBTextField sourcegraphUrlTextField;
    private final JBTextField accessTokenTextField;
    private final JBTextField defaultBranchNameTextField;
    private final JBTextField remoteUrlReplacementsTextField;
    private final JBCheckBox globbingCheckBox;

    public SettingsComponent() {
        JBLabel sourcegraphUrlLabel = new JBLabel("Sourcegraph URL:");
        sourcegraphUrlTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        sourcegraphUrlTextField.getEmptyText().setText("https://sourcegraph.example.com");
        sourcegraphUrlTextField.setToolTipText("If your company has your own Sourcegraph instance, set its URL here.\nThe default is \"https://sourcegraph.com\".");

        JBLabel accessTokenLabel = new JBLabel("Access token:");
        accessTokenTextField = new JBTextField();
        accessTokenTextField.getEmptyText().setText("Paste your access token here");
        accessTokenTextField.setToolTipText("Go to https://sourcegraph.example.com/users/{your-username}/settings/tokens to create a token.");

        JBLabel defaultBranchNameLabel = new JBLabel("Default branch name:");
        defaultBranchNameTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        defaultBranchNameTextField.getEmptyText().setText("main");
        defaultBranchNameTextField.setToolTipText("Usually \"main\" or \"master\", but can be any branch name");

        JBLabel remoteUrlReplacementsLabel = new JBLabel("Remote URL replacements:");
        remoteUrlReplacementsTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        remoteUrlReplacementsTextField.getEmptyText().setText("search1, replacement1, search2, replacement2, ...");
        remoteUrlReplacementsTextField.setToolTipText("You can replace specified strings in your repo's remote URL.\n" +
            "The format is \"search1, replacement1, search2, replacement2, ...\".\n" +
            "Use any number of search/replacement pairs.\n" +
            "Pairs will be replaced going from left to right.\n" +
            "The whitespace before and after the commas doesn't matter.");

        globbingCheckBox = new JBCheckBox("Enable globbing");

        panel = FormBuilder.createFormBuilder()
            .addLabeledComponent(sourcegraphUrlLabel, sourcegraphUrlTextField, 1, false)
            .addLabeledComponent(accessTokenLabel, accessTokenTextField, 1, false)
            .addLabeledComponent(defaultBranchNameLabel, defaultBranchNameTextField, 1, false)
            .addLabeledComponent(remoteUrlReplacementsLabel, remoteUrlReplacementsTextField, 1, false)
            .addComponent(globbingCheckBox, 1)
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

}
