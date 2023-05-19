package com.sourcegraph.cody.config;

import com.intellij.openapi.Disposable;
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
public class SettingsComponent implements Disposable {
    private final JPanel panel;
    private ButtonGroup instanceTypeButtonGroup;
    private JBLabel dotcomAccessTokenLinkComment;
    private JBTextField dotcomAccessTokenTextField;
    private JBTextField enterpriseUrlTextField;
    private JBTextField enterpriseAccessTokenTextField;
    private JBLabel userDocsLinkComment;
    private JBLabel enterpriseAccessTokenLinkComment;
    private JBTextField customRequestHeadersTextField;
    private JBTextField codebaseTextField;
    private JBCheckBox areChatPredictionsEnabledCheckBox;

    public JComponent getPreferredFocusedComponent() {
        return codebaseTextField;
    }

    public SettingsComponent() {
        JPanel userAuthenticationPanel = createAuthenticationPanel();
        JPanel navigationSettingsPanel = createOtherSettingsPanel();

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

        setDotcomSettingsEnabled(instanceType == InstanceType.DOTCOM);
        setEnterpriseSettingsEnabled(instanceType == InstanceType.ENTERPRISE);
    }

    @NotNull
    private JPanel createAuthenticationPanel() {

        // Create dotcom access token field
        JBLabel dotcomAccessTokenLabel = new JBLabel("Access token:");
        dotcomAccessTokenTextField = new JBTextField();
        dotcomAccessTokenTextField.getEmptyText().setText("Paste your access token here");
        addValidation(dotcomAccessTokenTextField, () ->
            isValidAccessToken(dotcomAccessTokenTextField.getText())
                ? null
                : new ValidationInfo("Invalid access token", dotcomAccessTokenTextField));
        dotcomAccessTokenLinkComment = new JBLabel("<html><body>Have no token yet? Create one <a href=\"https://sourcegraph.com/user/settings/tokens/new\">here</a>.</body></html>", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        //

        // Create URL field for the enterprise section
        JBLabel urlLabel = new JBLabel("Sourcegraph URL:");
        enterpriseUrlTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        enterpriseUrlTextField.getEmptyText().setText("https://sourcegraph.example.com");
        enterpriseUrlTextField.setToolTipText("The default is \"" + ConfigUtil.DOTCOM_URL + "\".");
        addValidation(enterpriseUrlTextField, () ->
            enterpriseUrlTextField.getText().length() == 0 ? new ValidationInfo("Missing URL", enterpriseUrlTextField)
                : (!JsonSchemaConfigurable.isValidURL(enterpriseUrlTextField.getText()) ? new ValidationInfo("This is an invalid URL", enterpriseUrlTextField)
                : null));
        addDocumentListener(enterpriseUrlTextField, e -> updateAccessTokenLinkCommentText());

        // Create enterprise access token field
        JBLabel enterpriseAccessTokenLabel = new JBLabel("Access token:");
        enterpriseAccessTokenTextField = new JBTextField();
        enterpriseAccessTokenTextField.getEmptyText().setText("Paste your access token here");
        addValidation(enterpriseAccessTokenTextField, () ->
            isValidAccessToken(enterpriseAccessTokenTextField.getText())
                ? null
                : new ValidationInfo("Invalid access token", enterpriseAccessTokenTextField));

        // Create comments
        userDocsLinkComment = new JBLabel("<html><body>You will need an access token to sign in. See our <a href=\"https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token\">user docs</a> for a video guide</body></html>", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        userDocsLinkComment.setBorder(JBUI.Borders.emptyLeft(10));
        enterpriseAccessTokenLinkComment = new JBLabel("", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        enterpriseAccessTokenLinkComment.setBorder(JBUI.Borders.emptyLeft(10));

        // Set up radio buttons
        ActionListener actionListener = event -> {
            String actionCommand = event.getActionCommand();
            setDotcomSettingsEnabled(actionCommand.equals(InstanceType.DOTCOM.name()));
            setEnterpriseSettingsEnabled(actionCommand.equals(InstanceType.ENTERPRISE.name()));
        };
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
        JBLabel dotComComment = new JBLabel("Cody for open source code is available to all users with a Sourcegraph.com account",
            UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
        dotComComment.setBorder(JBUI.Borders.emptyLeft(20));
        JPanel dotcomPanelContent = FormBuilder.createFormBuilder()
            .addLabeledComponent(dotcomAccessTokenLabel, dotcomAccessTokenTextField, 1)
            .addComponentToRightColumn(dotcomAccessTokenLinkComment, 1)
            .getPanel();
        dotcomPanelContent.setBorder(IdeBorderFactory.createEmptyBorder(JBUI.insets(1, 30, 0, 0)));
        JPanel dotComPanel = FormBuilder.createFormBuilder()
            .addComponent(sourcegraphDotComRadioButton, 1)
            .addComponentToRightColumn(dotComComment, 2)
            .addComponent(dotcomPanelContent, 1)
            .getPanel();
        JPanel enterprisePanelContent = FormBuilder.createFormBuilder()
            .addLabeledComponent(urlLabel, enterpriseUrlTextField, 1)
            .addTooltip("If your company uses a private Sourcegraph instance, set its URL here")
            .addLabeledComponent(enterpriseAccessTokenLabel, enterpriseAccessTokenTextField, 1)
            .addComponentToRightColumn(userDocsLinkComment, 1)
            .addComponentToRightColumn(enterpriseAccessTokenLinkComment, 1)
            .getPanel();
        enterprisePanelContent.setBorder(IdeBorderFactory.createEmptyBorder(JBUI.insets(1, 30, 0, 0)));
        JPanel enterprisePanel = FormBuilder.createFormBuilder()
            .addComponent(enterpriseInstanceRadioButton, 1)
            .addComponent(enterprisePanelContent, 1)
            .getPanel();

        // Create the "Request headers" text box
        JBLabel customRequestHeadersLabel = new JBLabel("Custom headers:");
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
        userAuthenticationPanel.setBorder(IdeBorderFactory.createTitledBorder("Authentication", true, JBUI.insetsTop(8)));

        return userAuthenticationPanel;
    }

    @NotNull
    public String getDotcomAccessToken() {
        return dotcomAccessTokenTextField.getText();
    }

    public void setDotcomAccessToken(@NotNull String value) {
        dotcomAccessTokenTextField.setText(value);
    }

    @NotNull
    public String getEnterpriseUrl() {
        return enterpriseUrlTextField.getText();
    }

    public void setEnterpriseUrl(@Nullable String value) {
        enterpriseUrlTextField.setText(value != null ? value : "");
    }

    @NotNull
    public String getEnterpriseAccessToken() {
        return enterpriseAccessTokenTextField.getText();
    }

    public void setEnterpriseAccessToken(@NotNull String value) {
        enterpriseAccessTokenTextField.setText(value);
    }

    @NotNull
    public String getCustomRequestHeaders() {
        return customRequestHeadersTextField.getText();
    }

    public void setCustomRequestHeaders(@NotNull String customRequestHeaders) {
        this.customRequestHeadersTextField.setText(customRequestHeaders);
    }

    @NotNull
    public String getCodebase() {
        return codebaseTextField.getText();
    }

    public void setCodebase(@NotNull String value) {
        codebaseTextField.setText(value);
    }

    public boolean areChatPredictionsEnabled() {
        return areChatPredictionsEnabledCheckBox != null && areChatPredictionsEnabledCheckBox.isSelected();
    }

    public void setAreChatPredictionsEnabled(boolean value) {
        if (areChatPredictionsEnabledCheckBox != null) {
            areChatPredictionsEnabledCheckBox.setSelected(value);
        }
    }

    private void setDotcomSettingsEnabled(boolean enable) {
        dotcomAccessTokenTextField.setEnabled(enable);
        dotcomAccessTokenLinkComment.setEnabled(enable);
        dotcomAccessTokenLinkComment.setCopyable(enable);
    }

    private void setEnterpriseSettingsEnabled(boolean enable) {
        enterpriseUrlTextField.setEnabled(enable);
        enterpriseAccessTokenTextField.setEnabled(enable);
        userDocsLinkComment.setEnabled(enable);
        userDocsLinkComment.setCopyable(enable);
        enterpriseAccessTokenLinkComment.setEnabled(enable);
        enterpriseAccessTokenLinkComment.setCopyable(enable);
    }

    @Override
    public void dispose() {
        instanceTypeButtonGroup = null;
        enterpriseUrlTextField = null;
        enterpriseAccessTokenTextField = null;
        customRequestHeadersTextField = null;
        codebaseTextField = null;
        areChatPredictionsEnabledCheckBox = null;
        userDocsLinkComment = null;
        enterpriseAccessTokenLinkComment = null;

    }

    public enum InstanceType {
        DOTCOM,
        ENTERPRISE,
    }

    private void addValidation(@NotNull JTextComponent component, @NotNull Supplier<ValidationInfo> validator) {
        new ComponentValidator(this).withValidator(validator).installOn(component);
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
        String baseUrl = enterpriseUrlTextField.getText();
        String settingsUrl = (baseUrl.endsWith("/") ? baseUrl : baseUrl + "/") + "settings";
        enterpriseAccessTokenLinkComment.setText(isUrlValid(baseUrl)
            ? "<html><body>or go to <a href=\"" + settingsUrl + "\">" + settingsUrl + "</a> | \"Access tokens\" to create one.</body></html>"
            : "");
    }

    private boolean isValidAccessToken(@NotNull String accessToken) {
        return accessToken.isEmpty() ||
            accessToken.length() == 40 ||
            (accessToken.startsWith("sgp_") && accessToken.length() == 44);
    }

    private boolean isUrlValid(@NotNull String url) {
        return JsonSchemaConfigurable.isValidURL(url);
    }

    @NotNull
    private JPanel createOtherSettingsPanel() {
        JBLabel codebaseLabel = new JBLabel("Codebase:");
        codebaseTextField = new JBTextField();
        //noinspection DialogTitleCapitalization
        codebaseTextField.getEmptyText().setText("github.com/sourcegraph/sourcegraph");

        //// Always disabled for now
        //areChatPredictionsEnabledCheckBox = new JBCheckBox("Experimental: Chat predictions");
        //areChatPredictionsEnabledCheckBox.setEnabled(false);

        //noinspection DialogTitleCapitalization
        JPanel otherSettingsPanel = FormBuilder.createFormBuilder()
            .addLabeledComponent(codebaseLabel, codebaseTextField)
            .addTooltip("The name of the embedded repository that Cody will use to gather context")
            .addTooltip("for its responses. This is automatically inferred from your Git metadata,")
            .addTooltip("but you can use this option if you need to override the default.")
            //.addComponent(areChatPredictionsEnabledCheckBox, 10)
            //.addTooltip("Adds suggestions of possible relevant messages in the chat window")
            .getPanel();
        otherSettingsPanel.setBorder(IdeBorderFactory.createTitledBorder("Other Settings", true, JBUI.insetsTop(8)));
        return otherSettingsPanel;
    }
}
