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
import org.jetbrains.annotations.Nullable;

import javax.swing.*;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.JTextComponent;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.util.Enumeration;
import java.util.function.Consumer;
import java.util.function.Supplier;

/**
 * Supports creating and managing a {@link JPanel} for the Settings Dialog.
 */
public class SettingsComponent {
    private final Project project;
    private final JPanel panel;
    private ButtonGroup instanceTypeButtonGroup;
    private JBTextField urlTextField;
    private JBTextField accessTokenTextField;
    private JBLabel userDocsLinkComment;
    private JBLabel accessTokenLinkComment;
    private JBTextField customRequestHeadersTextField;
    private JBTextField defaultBranchNameTextField;
    private JBTextField remoteUrlReplacementsTextField;
    private JBCheckBox globbingCheckBox;
    private JBCheckBox isUrlNotificationDismissedCheckBox;

    public JComponent getPreferredFocusedComponent() {
        return defaultBranchNameTextField;
    }

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

    @NotNull
    public InstanceType getInstanceType() {
        return instanceTypeButtonGroup.getSelection().getActionCommand().equals(InstanceType.DOTCOM.name()) ? InstanceType.DOTCOM : InstanceType.ENTERPRISE;
    }

    public void setInstanceType(@NotNull InstanceType instanceType) {
        for (Enumeration<AbstractButton> buttons = instanceTypeButtonGroup.getElements(); buttons.hasMoreElements(); ) {
            AbstractButton button = buttons.nextElement();

            button.setSelected(button.getActionCommand().equals(instanceType.name()));
        }

        setEnterpriseSettingsEnabled(instanceType == InstanceType.ENTERPRISE);
    }

    @NotNull
    private JPanel createAuthenticationPanel() {
        // Create URL field for the enterprise section
        JBLabel urlLabel = new JBLabel("Sourcegraph URL:");
        urlTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        urlTextField.getEmptyText().setText("https://sourcegraph.example.com");
        urlTextField.setToolTipText("The default is \"" + ConfigUtil.DOTCOM_URL + "\".");
        addValidation(urlTextField, () ->
            urlTextField.getText().length() == 0 ? new ValidationInfo("Missing URL", urlTextField)
                : (!JsonSchemaConfigurable.isValidURL(urlTextField.getText()) ? new ValidationInfo("This is an invalid URL", urlTextField)
                : null));
        addDocumentListener(urlTextField, e -> updateAccessTokenLinkCommentText());

        // Create access token field
        JBLabel accessTokenLabel = new JBLabel("Access token:");
        accessTokenTextField = new JBTextField();
        accessTokenTextField.getEmptyText().setText("Paste your access token here");
        addValidation(accessTokenTextField, () ->
            (accessTokenTextField.getText().length() > 0 && accessTokenTextField.getText().length() != 40)
                ? new ValidationInfo("Invalid access token", accessTokenTextField)
                : null);

        // Create comments
        userDocsLinkComment = new JBLabel("<html><body>You will need an access token to sign in. See our <a href=\"https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token\">user docs</a> for a video guide</body></html>", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        userDocsLinkComment.setBorder(JBUI.Borders.emptyLeft(10));
        accessTokenLinkComment = new JBLabel("", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        accessTokenLinkComment.setBorder(JBUI.Borders.emptyLeft(10));

        // Set up radio buttons
        ActionListener actionListener = event -> setEnterpriseSettingsEnabled(event.getActionCommand().equals(InstanceType.ENTERPRISE.name()));
        JRadioButton sourcegraphDotComRadioButton = new JRadioButton("Use sourcegraph.com");
        sourcegraphDotComRadioButton.setMnemonic(KeyEvent.VK_C);
        sourcegraphDotComRadioButton.setActionCommand(InstanceType.DOTCOM.name());
        sourcegraphDotComRadioButton.addActionListener(actionListener);
        JRadioButton enterpriseInstanceRadioButton = new JRadioButton("Use an enterprise instance");
        enterpriseInstanceRadioButton.setMnemonic(KeyEvent.VK_E);
        enterpriseInstanceRadioButton.setActionCommand(InstanceType.ENTERPRISE.name());
        enterpriseInstanceRadioButton.addActionListener(actionListener);
        instanceTypeButtonGroup = new ButtonGroup();
        instanceTypeButtonGroup.add(sourcegraphDotComRadioButton);
        instanceTypeButtonGroup.add(enterpriseInstanceRadioButton);

        // Assemble the two main panels
        JBLabel dotComComment = new JBLabel("Use sourcegraph.com to search public code", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        dotComComment.setBorder(JBUI.Borders.emptyLeft(20));
        JPanel dotComPanel = FormBuilder.createFormBuilder()
            .addComponent(sourcegraphDotComRadioButton, 1)
            .addComponentToRightColumn(dotComComment, 2)
            .getPanel();
        JPanel enterprisePanelContent = FormBuilder.createFormBuilder()
            .addLabeledComponent(urlLabel, urlTextField, 1)
            .addTooltip("If your company uses a private Sourcegraph instance, set its URL here")
            .addLabeledComponent(accessTokenLabel, accessTokenTextField, 1)
            .addComponentToRightColumn(userDocsLinkComment, 1)
            .addComponentToRightColumn(accessTokenLinkComment, 1)
            .getPanel();
        enterprisePanelContent.setBorder(IdeBorderFactory.createEmptyBorder(JBUI.insets(1, 30, 0, 0)));
        JPanel enterprisePanel = FormBuilder.createFormBuilder()
            .addComponent(enterpriseInstanceRadioButton, 1)
            .addComponent(enterprisePanelContent, 1)
            .getPanel();

        // Create the "Request headers" text box
        JBLabel customRequestHeadersLabel = new JBLabel("Custom request headers:");
        customRequestHeadersTextField = new JBTextField();
        customRequestHeadersTextField.getEmptyText().setText("Client-ID, client-one, X-Extra, some metadata");
        customRequestHeadersTextField.setToolTipText("You can even overwrite \"Authorization\" that Access token sets above.");
        addValidation(customRequestHeadersTextField, () -> {
            if (customRequestHeadersTextField.getText().length() == 0) {
                return null;
            }
            String[] pairs = customRequestHeadersTextField.getText().split(",");
            if (pairs.length % 2 != 0) {
                return new ValidationInfo("Must be a comma-separated list of pairs", customRequestHeadersTextField);
            }

            for (int i = 0; i < pairs.length; i += 2) {
                String headerName = pairs[i].trim();
                if (!headerName.matches("[\\w-]+")) {
                    return new ValidationInfo("Invalid HTTP header name: " + headerName, customRequestHeadersTextField);
                }
            }
            return null;
        });

        // Assemble the main panel
        JPanel userAuthenticationPanel = FormBuilder.createFormBuilder()
            .addComponent(dotComPanel)
            .addComponent(enterprisePanel, 5)
            .addLabeledComponent(customRequestHeadersLabel, customRequestHeadersTextField)
            .addTooltip("Any custom headers to send with every request to Sourcegraph.")
            .addTooltip("Use any number of pairs: \"header1, value1, header2, value2, ...\".")
            .addTooltip("Whitespace around commas doesn't matter.")
            .getPanel();
        userAuthenticationPanel.setBorder(IdeBorderFactory.createTitledBorder("User Authentication", true, JBUI.insetsTop(8)));

        return userAuthenticationPanel;
    }

    @NotNull
    public String getEnterpriseUrl() {
        return urlTextField.getText();
    }

    public void setEnterpriseUrl(@Nullable String value) {
        urlTextField.setText(value != null ? value : "");
    }

    @NotNull
    public String getAccessToken() {
        return accessTokenTextField.getText();
    }

    public void setAccessToken(@NotNull String value) {
        accessTokenTextField.setText(value);
    }

    @NotNull
    public String getCustomRequestHeaders() {
        return customRequestHeadersTextField.getText();
    }

    public void setCustomRequestHeaders(@NotNull String customRequestHeaders) {
        this.customRequestHeadersTextField.setText(customRequestHeaders);
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

    private void setEnterpriseSettingsEnabled(boolean enable) {
        urlTextField.setEnabled(enable);
        accessTokenTextField.setEnabled(enable);
        userDocsLinkComment.setEnabled(enable);
        userDocsLinkComment.setCopyable(enable);
        accessTokenLinkComment.setEnabled(enable);
        accessTokenLinkComment.setCopyable(enable);
    }

    public enum InstanceType {
        DOTCOM,
        ENTERPRISE,
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
        String baseUrl = urlTextField.getText();
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
