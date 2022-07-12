package com.sourcegraph.config;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.ComponentValidator;
import com.intellij.openapi.ui.ValidationInfo;
import com.intellij.ui.IdeBorderFactory;
import com.intellij.ui.components.JBCheckBox;
import com.intellij.ui.components.JBLabel;
import com.intellij.ui.components.JBTextField;
import com.intellij.util.ui.FormBuilder;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.UIUtil;
import com.jetbrains.jsonSchema.settings.mappings.JsonSchemaConfigurable;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.JTextComponent;
import java.util.function.Consumer;
import java.util.function.Supplier;

/**
 * Supports creating and managing a {@link JPanel} for the Settings Dialog.
 */
public class SettingsComponent {
    private final Project project;
    private final JPanel panel;
    private JBTextField sourcegraphUrlTextField;
    private JBTextField accessTokenTextField;
    private JBTextField defaultBranchNameTextField;
    private JBTextField remoteUrlReplacementsTextField;
    private JBCheckBox globbingCheckBox;
    private JBCheckBox isUrlNotificationDismissedCheckBox;
    private JBLabel accessTokenLinkComment;

    public SettingsComponent(@NotNull Project project) {
        this.project = project;
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
        JBLabel urlLabel = new JBLabel("Sourcegraph URL:");
        sourcegraphUrlTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        sourcegraphUrlTextField.getEmptyText().setText("https://sourcegraph.example.com");
        sourcegraphUrlTextField.setToolTipText("The default is \"https://sourcegraph.com\".");

        addValidation(sourcegraphUrlTextField, () ->
            sourcegraphUrlTextField.getText().length() == 0 ? new ValidationInfo("Missing URL", sourcegraphUrlTextField)
                : (!JsonSchemaConfigurable.isValidURL(sourcegraphUrlTextField.getText()) ? new ValidationInfo("This is an invalid URL", sourcegraphUrlTextField)
                : null));
        addDocumentListener(sourcegraphUrlTextField, e -> updateAccessTokenLinkCommentText());

        JBLabel accessTokenLabel = new JBLabel("Access token:");
        accessTokenTextField = new JBTextField();
        accessTokenTextField.getEmptyText().setText("Paste your access token here");
        addValidation(accessTokenTextField, () ->
            (accessTokenTextField.getText().length() > 0 && accessTokenTextField.getText().length() != 40)
                ? new ValidationInfo("Invalid access token", accessTokenTextField)
                : null);

        JBLabel userDocsLinkComment = new JBLabel("<html><body>You might need an access token to sign in. See our <a href=\"https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token\">user docs</a> for a video guide,</body></html>", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        userDocsLinkComment.setBorder(JBUI.Borders.emptyLeft(10));
        userDocsLinkComment.setCopyable(true);
        accessTokenLinkComment = new JBLabel("", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        accessTokenLinkComment.setBorder(JBUI.Borders.emptyLeft(10));
        accessTokenLinkComment.setCopyable(true);

        JPanel userAuthenticationPanel = FormBuilder.createFormBuilder()
            .addLabeledComponent(urlLabel, sourcegraphUrlTextField)
            .addTooltip("If your company has your own Sourcegraph instance, set its URL here")
            .addLabeledComponent(accessTokenLabel, accessTokenTextField)
            .addComponentToRightColumn(userDocsLinkComment, 1)
            .addComponentToRightColumn(accessTokenLinkComment, 1)
            .getPanel();
        userAuthenticationPanel.setBorder(IdeBorderFactory.createTitledBorder("User Authentication", true, JBUI.insetsTop(8)));

        return userAuthenticationPanel;
    }

    private void addValidation(@NotNull JTextComponent component, @NotNull Supplier<ValidationInfo> validator) {
        new ComponentValidator(project).withValidator(validator).installOn(component);
        addDocumentListener(component, e -> ComponentValidator.getInstance(component).ifPresent(ComponentValidator::revalidate));
    }

    private void addDocumentListener(@NotNull JTextComponent textComponent, @NotNull Consumer<ComponentValidator> function) {
        textComponent.getDocument().addDocumentListener(new DocumentListener() {
            @Override
            public void insertUpdate(DocumentEvent e) {
                ComponentValidator.getInstance(textComponent).ifPresent(function);
            }

            @Override
            public void removeUpdate(DocumentEvent e) {
                ComponentValidator.getInstance(textComponent).ifPresent(function);
            }

            @Override
            public void changedUpdate(DocumentEvent e) {
                ComponentValidator.getInstance(textComponent).ifPresent(function);
            }
        });
    }

    private void updateAccessTokenLinkCommentText() {
        String baseUrl = sourcegraphUrlTextField.getText();
        String settingsUrl = (baseUrl.endsWith("/") ? baseUrl : baseUrl + "/") + "settings";
        accessTokenLinkComment.setText(isUrlValid(baseUrl)
            ? "<html><body>or go to <a href=\"" + settingsUrl + "\">" + settingsUrl + "</a> | \"Access tokens\" to create one.</body></html>"
            : "");
    }

    private boolean isUrlValid(@NotNull String url) {
        return JsonSchemaConfigurable.isValidURL(url);
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
        addValidation(remoteUrlReplacementsTextField, () ->
            (remoteUrlReplacementsTextField.getText().length() > 0 && remoteUrlReplacementsTextField.getText().split(",").length % 2 != 0)
                ? new ValidationInfo("Must be a comma-separated list of pairs", remoteUrlReplacementsTextField)
                : null);

        globbingCheckBox = new JBCheckBox("Enable globbing");
        isUrlNotificationDismissedCheckBox = new JBCheckBox("Do not show the \"No Sourcegraph URL set\" notification for this project");

        JPanel navigationSettingsPanel = FormBuilder.createFormBuilder()
            .addLabeledComponent(defaultBranchNameLabel, defaultBranchNameTextField)
            .addTooltip("The branch to use if the current branch is not yet pushed")
            .addLabeledComponent(remoteUrlReplacementsLabel, remoteUrlReplacementsTextField)
            .addTooltip("You can replace specified strings in your repo's remote URL.")
            .addTooltip("Use any number of pairs: \"search1, replacement1, search2, replacement2, ...\".")
            .addTooltip("Pairs are replaced from left to right. Whitespace around commas doesn't matter.")
            .addComponent(globbingCheckBox)
            .addComponent(isUrlNotificationDismissedCheckBox, 10)
            .getPanel();
        navigationSettingsPanel.setBorder(IdeBorderFactory.createTitledBorder("Navigation Settings", true, JBUI.insetsTop(8)));
        return navigationSettingsPanel;
    }
}
