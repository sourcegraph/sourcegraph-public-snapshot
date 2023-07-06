package com.sourcegraph.config;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.ComponentValidator;
import com.intellij.openapi.ui.ValidationInfo;
import com.intellij.ui.IdeBorderFactory;
import com.intellij.ui.components.ActionLink;
import com.intellij.ui.components.JBCheckBox;
import com.intellij.ui.components.JBLabel;
import com.intellij.ui.components.JBTextField;
import com.intellij.util.ui.FormBuilder;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.UIUtil;
import com.jetbrains.jsonSchema.settings.mappings.JsonSchemaConfigurable;
import com.sourcegraph.cody.localapp.LocalAppManager;
import com.sourcegraph.common.AuthorizationUtil;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.util.Enumeration;
import java.util.Optional;
import java.util.function.Consumer;
import java.util.function.Supplier;
import javax.swing.*;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.JTextComponent;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/** Supports creating and managing a {@link JPanel} for the Settings Dialog. */
public class SettingsComponent {
  private final Project project;
  private final JPanel panel;
  private ButtonGroup instanceTypeButtonGroup;
  private JBTextField urlTextField;
  private JBTextField enterpriseAccessTokenTextField;
  private JBTextField dotComAccessTokenTextField;
  private JBLabel userDocsLinkComment;
  private JBLabel enterpriseAccessTokenLinkComment;
  private JBTextField customRequestHeadersTextField;
  private JBTextField defaultBranchNameTextField;
  private JBTextField remoteUrlReplacementsTextField;
  private JBCheckBox isUrlNotificationDismissedCheckBox;
  private JBCheckBox isCodyAutoCompleteEnabledCheckBox;

  private JButton testCodyAppConnectionButton;
  private JLabel testCodyAppConnectionLabel;

  public JComponent getPreferredFocusedComponent() {
    return defaultBranchNameTextField;
  }

  public SettingsComponent(@NotNull Project project) {
    this.project = project;
    JPanel userAuthenticationPanel = createAuthenticationPanel();
    JPanel navigationSettingsPanel = createNavigationSettingsPanel();
    JPanel codySettingsPanel = createCodySettingsPanel();

    panel =
        FormBuilder.createFormBuilder()
            .addComponent(userAuthenticationPanel)
            .addComponent(navigationSettingsPanel)
            .addComponent(codySettingsPanel)
            .addComponentFillVertically(new JPanel(), 0)
            .getPanel();
  }

  public JPanel getPanel() {
    return panel;
  }

  public static InstanceType getDefaultInstanceType() {
    return LocalAppManager.isLocalAppInstalled() && LocalAppManager.isPlatformSupported()
        ? InstanceType.LOCAL_APP
        : InstanceType.DOTCOM;
  }

  @NotNull
  public InstanceType getInstanceType() {
    return InstanceType.optionalValueOf(instanceTypeButtonGroup.getSelection().getActionCommand())
        .orElse(getDefaultInstanceType());
  }

  public void setInstanceType(@NotNull InstanceType instanceType) {
    for (Enumeration<AbstractButton> buttons = instanceTypeButtonGroup.getElements();
        buttons.hasMoreElements(); ) {
      AbstractButton button = buttons.nextElement();

      button.setSelected(button.getActionCommand().equals(instanceType.name()));
    }

    setInstanceSettingsEnabled(instanceType);
  }

  private ActionLink simpleActionLink(String text, Runnable runnable) {
    ActionLink actionLink =
        new ActionLink(
            text,
            e -> {
              runnable.run();
            });
    actionLink.setFont(UIUtil.getLabelFont(UIUtil.FontSize.SMALL));
    return actionLink;
  }

  @NotNull
  private JPanel createAuthenticationPanel() {
    // Create URL field for the enterprise section
    JBLabel urlLabel = new JBLabel("Sourcegraph URL:");
    urlTextField = new JBTextField();
    //noinspection DialogTitleCapitalization
    urlTextField.getEmptyText().setText("https://sourcegraph.example.com");
    urlTextField.setToolTipText("The default is \"" + ConfigUtil.DOTCOM_URL + "\".");
    addValidation(
        urlTextField,
        () ->
            urlTextField.getText().length() == 0
                ? new ValidationInfo("Missing URL", urlTextField)
                : (!JsonSchemaConfigurable.isValidURL(urlTextField.getText())
                    ? new ValidationInfo("This is an invalid URL", urlTextField)
                    : null));
    addDocumentListener(urlTextField, e -> updateAccessTokenLinkCommentText());

    // Create access token field
    JBLabel accessTokenLabel = new JBLabel("Access token:");
    enterpriseAccessTokenTextField = new JBTextField();
    enterpriseAccessTokenTextField.getEmptyText().setText("Paste your access token here");
    addValidation(
        enterpriseAccessTokenTextField,
        () ->
            !AuthorizationUtil.isValidAccessToken(enterpriseAccessTokenTextField.getText())
                ? new ValidationInfo("Invalid access token", enterpriseAccessTokenTextField)
                : null);

    // Create access token field
    JBLabel dotComAccessTokenComment =
        new JBLabel(
                "(optional) To use Cody, you will need an access token to sign in.",
                UIUtil.ComponentStyle.SMALL,
                UIUtil.FontColor.BRIGHTER)
            .withBorder(JBUI.Borders.emptyLeft(10));
    JBLabel dotComAccessTokenLabel = new JBLabel("Access token:");
    dotComAccessTokenTextField = new JBTextField();
    dotComAccessTokenTextField.getEmptyText().setText("Paste your access token here");
    addValidation(
        dotComAccessTokenTextField,
        () ->
            !AuthorizationUtil.isValidAccessToken(dotComAccessTokenTextField.getText())
                ? new ValidationInfo("Invalid access token", dotComAccessTokenTextField)
                : null);

    // Create comments
    userDocsLinkComment =
        new JBLabel(
            "<html><body>You will need an access token to sign in. See our <a href=\"https://docs.sourcegraph.com/cli/how-tos/creating_an_access_token\">user docs</a> for a video guide</body></html>",
            UIUtil.ComponentStyle.SMALL,
            UIUtil.FontColor.BRIGHTER);
    userDocsLinkComment.setBorder(JBUI.Borders.emptyLeft(10));
    enterpriseAccessTokenLinkComment =
        new JBLabel("", UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
    enterpriseAccessTokenLinkComment.setBorder(JBUI.Borders.emptyLeft(10));

    // Set up radio buttons
    ActionListener actionListener =
        event ->
            setInstanceSettingsEnabled(
                InstanceType.optionalValueOf(event.getActionCommand())
                    .orElse(getDefaultInstanceType()));
    boolean isLocalAppInstalled = LocalAppManager.isLocalAppInstalled();
    boolean isLocalAppAccessTokenConfigured = LocalAppManager.getLocalAppAccessToken().isPresent();
    boolean isLocalAppRunning = LocalAppManager.isLocalAppRunning();
    boolean isLocalAppPlatformSupported = LocalAppManager.isPlatformSupported();
    JRadioButton codyAppRadioButton = new JRadioButton("Use the local Cody App");
    codyAppRadioButton.setMnemonic(KeyEvent.VK_A);
    codyAppRadioButton.setActionCommand(InstanceType.LOCAL_APP.name());
    codyAppRadioButton.addActionListener(actionListener);
    codyAppRadioButton.setEnabled(isLocalAppInstalled && isLocalAppAccessTokenConfigured);
    JRadioButton dotcomRadioButton = new JRadioButton("Use sourcegraph.com");
    dotcomRadioButton.setMnemonic(KeyEvent.VK_C);
    dotcomRadioButton.setActionCommand(InstanceType.DOTCOM.name());
    dotcomRadioButton.addActionListener(actionListener);
    JRadioButton enterpriseInstanceRadioButton = new JRadioButton("Use an enterprise instance");
    enterpriseInstanceRadioButton.setMnemonic(KeyEvent.VK_E);
    enterpriseInstanceRadioButton.setActionCommand(InstanceType.ENTERPRISE.name());
    enterpriseInstanceRadioButton.addActionListener(actionListener);
    instanceTypeButtonGroup = new ButtonGroup();
    instanceTypeButtonGroup.add(codyAppRadioButton);
    instanceTypeButtonGroup.add(dotcomRadioButton);
    instanceTypeButtonGroup.add(enterpriseInstanceRadioButton);

    // Assemble the three main panels String platformName =
    String platformName =
        Optional.ofNullable(System.getProperty("os.name")).orElse("Your platform");
    String codyAppCommentText =
        isLocalAppPlatformSupported
            ? "Use Sourcegraph through Cody App."
            : platformName
                + " is not yet supported by the Cody App. Keep an eye on future updates!";
    JBLabel codyAppComment =
        new JBLabel(codyAppCommentText, UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
    codyAppComment.setBorder(JBUI.Borders.emptyLeft(20));
    boolean shouldShowInstallLocalAppLink = !isLocalAppInstalled && isLocalAppPlatformSupported;
    JLabel installLocalAppComment =
        new JBLabel(
            "Cody App wasn't detected on this system, it seems it hasn't been installed yet.",
            UIUtil.ComponentStyle.SMALL,
            UIUtil.FontColor.BRIGHTER);
    installLocalAppComment.setVisible(shouldShowInstallLocalAppLink);
    installLocalAppComment.setBorder(JBUI.Borders.emptyLeft(20));
    ActionLink installLocalAppLink =
        simpleActionLink("Install Cody App...", LocalAppManager::browseLocalAppInstallPage);
    installLocalAppLink.setVisible(shouldShowInstallLocalAppLink);
    installLocalAppLink.setBorder(JBUI.Borders.emptyLeft(20));
    boolean shouldShowRunLocalAppLink = isLocalAppInstalled && !isLocalAppRunning;
    ActionLink runLocalAppLink = simpleActionLink("Run Cody App...", LocalAppManager::runLocalApp);
    runLocalAppLink.setVisible(shouldShowRunLocalAppLink);
    runLocalAppLink.setBorder(JBUI.Borders.emptyLeft(20));
    JLabel runLocalAppComment =
        new JBLabel(
            "Cody App seems to be installed, but it's not running, currently.",
            UIUtil.ComponentStyle.SMALL,
            UIUtil.FontColor.BRIGHTER);
    runLocalAppComment.setVisible(shouldShowRunLocalAppLink);
    runLocalAppComment.setBorder(JBUI.Borders.emptyLeft(20));
    JPanel codyAppPanel =
        FormBuilder.createFormBuilder()
            .addComponent(codyAppRadioButton, 1)
            .addComponent(codyAppComment, 2)
            .addComponent(installLocalAppComment, 2)
            .addComponent(installLocalAppLink, 2)
            .addComponent(runLocalAppComment, 2)
            .addComponent(runLocalAppLink, 2)
            .getPanel();
    JBLabel dotComComment =
        new JBLabel(
            "Use sourcegraph.com to search public code",
            UIUtil.ComponentStyle.SMALL,
            UIUtil.FontColor.BRIGHTER);
    dotComComment.setBorder(JBUI.Borders.emptyLeft(20));
    JPanel dotComPanelContent =
        FormBuilder.createFormBuilder()
            .addLabeledComponent(dotComAccessTokenLabel, dotComAccessTokenTextField, 1)
            .addComponentToRightColumn(dotComAccessTokenComment, 1)
            .getPanel();
    dotComPanelContent.setBorder(IdeBorderFactory.createEmptyBorder(JBUI.insets(1, 30, 0, 0)));
    JPanel dotComPanel =
        FormBuilder.createFormBuilder()
            .addComponent(dotcomRadioButton, 1)
            .addComponent(dotComComment, 2)
            .addComponent(dotComPanelContent, 1)
            .getPanel();
    JPanel enterprisePanelContent =
        FormBuilder.createFormBuilder()
            .addLabeledComponent(urlLabel, urlTextField, 1)
            .addTooltip("If your company uses a Sourcegraph enterprise instance, set its URL here")
            .addLabeledComponent(accessTokenLabel, enterpriseAccessTokenTextField, 1)
            .addComponentToRightColumn(userDocsLinkComment, 1)
            .addComponentToRightColumn(enterpriseAccessTokenLinkComment, 1)
            .getPanel();
    enterprisePanelContent.setBorder(IdeBorderFactory.createEmptyBorder(JBUI.insets(1, 30, 0, 0)));
    JPanel enterprisePanel =
        FormBuilder.createFormBuilder()
            .addComponent(enterpriseInstanceRadioButton, 1)
            .addComponent(enterprisePanelContent, 1)
            .getPanel();

    // Create the "Request headers" text box
    JBLabel customRequestHeadersLabel = new JBLabel("Custom request headers:");
    customRequestHeadersTextField = new JBTextField();
    customRequestHeadersTextField
        .getEmptyText()
        .setText("Client-ID, client-one, X-Extra, some metadata");
    customRequestHeadersTextField.setToolTipText(
        "You can even overwrite \"Authorization\" that Access token sets above.");
    addValidation(
        customRequestHeadersTextField,
        () -> {
          if (customRequestHeadersTextField.getText().length() == 0) {
            return null;
          }
          String[] pairs = customRequestHeadersTextField.getText().split(",");
          if (pairs.length % 2 != 0) {
            return new ValidationInfo(
                "Must be a comma-separated list of string pairs", customRequestHeadersTextField);
          }

          for (int i = 0; i < pairs.length; i += 2) {
            String headerName = pairs[i].trim();
            if (!headerName.matches("[\\w-]+")) {
              return new ValidationInfo(
                  "Invalid HTTP header name: " + headerName, customRequestHeadersTextField);
            }
          }
          return null;
        });

    // Assemble the main panel
    JPanel userAuthenticationPanel =
        FormBuilder.createFormBuilder()
            .addComponent(codyAppPanel)
            .addComponent(dotComPanel, 5)
            .addComponent(enterprisePanel, 5)
            .addLabeledComponent(customRequestHeadersLabel, customRequestHeadersTextField)
            .addTooltip("Any custom headers to send with every request to Sourcegraph.")
            .addTooltip("Use any number of pairs: \"header1, value1, header2, value2, ...\".")
            .addTooltip("Whitespace around commas doesn't matter.")
            .getPanel();
    userAuthenticationPanel.setBorder(
        IdeBorderFactory.createTitledBorder("User Authentication", true, JBUI.insetsTop(8)));

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
  public String getDotComAccessToken() {
    return dotComAccessTokenTextField.getText();
  }

  public void setDotComAccessToken(@NotNull String value) {
    dotComAccessTokenTextField.setText(value);
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

  public boolean isUrlNotificationDismissed() {
    return isUrlNotificationDismissedCheckBox.isSelected();
  }

  public void setUrlNotificationDismissedEnabled(boolean value) {
    isUrlNotificationDismissedCheckBox.setSelected(value);
  }

  public boolean isCodyAutoCompleteEnabled() {
    return isCodyAutoCompleteEnabledCheckBox.isSelected();
  }

  public void setCodyAutoCompleteEnabled(boolean value) {
    isCodyAutoCompleteEnabledCheckBox.setSelected(value);
  }

  private void setInstanceSettingsEnabled(@NotNull InstanceType instanceType) {
    // enterprise stuff
    boolean isEnterprise = instanceType == InstanceType.ENTERPRISE;
    urlTextField.setEnabled(isEnterprise);
    enterpriseAccessTokenTextField.setEnabled(isEnterprise);
    userDocsLinkComment.setEnabled(isEnterprise);
    userDocsLinkComment.setCopyable(isEnterprise);
    enterpriseAccessTokenLinkComment.setEnabled(isEnterprise);
    enterpriseAccessTokenLinkComment.setCopyable(isEnterprise);

    // dotCom stuff
    boolean isDotCom = instanceType == InstanceType.DOTCOM;
    dotComAccessTokenTextField.setEnabled(isDotCom);
  }

  public enum InstanceType {
    DOTCOM,
    ENTERPRISE,
    LOCAL_APP;

    @NotNull
    public static Optional<InstanceType> optionalValueOf(@NotNull String name) {
      try {
        return Optional.of(InstanceType.valueOf(name));
      } catch (IllegalArgumentException e) {
        return Optional.empty();
      }
    }
  }

  private void addValidation(
      @NotNull JTextComponent component, @NotNull Supplier<ValidationInfo> validator) {
    new ComponentValidator(project).withValidator(validator).installOn(component);
    addDocumentListener(
        component,
        e -> ComponentValidator.getInstance(component).ifPresent(ComponentValidator::revalidate));
  }

  private void addDocumentListener(
      @NotNull JTextComponent textComponent, @NotNull Consumer<ComponentValidator> function) {
    textComponent
        .getDocument()
        .addDocumentListener(
            new DocumentListener() {
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
    enterpriseAccessTokenLinkComment.setText(
        isUrlValid(baseUrl)
            ? "<html><body>or go to <a href=\""
                + settingsUrl
                + "\">"
                + settingsUrl
                + "</a> | \"Access tokens\" to create one.</body></html>"
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
    defaultBranchNameTextField.setToolTipText(
        "Usually \"main\" or \"master\", but can be any name");

    JBLabel remoteUrlReplacementsLabel = new JBLabel("Remote URL replacements:");
    remoteUrlReplacementsTextField = new JBTextField();
    //noinspection DialogTitleCapitalization
    remoteUrlReplacementsTextField
        .getEmptyText()
        .setText("search1, replacement1, search2, replacement2, ...");
    addValidation(
        remoteUrlReplacementsTextField,
        () ->
            (remoteUrlReplacementsTextField.getText().length() > 0
                    && remoteUrlReplacementsTextField.getText().split(",").length % 2 != 0)
                ? new ValidationInfo(
                    "Must be a comma-separated list of pairs", remoteUrlReplacementsTextField)
                : null);

    isUrlNotificationDismissedCheckBox =
        new JBCheckBox("Do not show the \"No Sourcegraph URL set\" notification for this project");

    JPanel navigationSettingsPanel =
        FormBuilder.createFormBuilder()
            .addLabeledComponent(defaultBranchNameLabel, defaultBranchNameTextField)
            .addTooltip("The branch to use if the current branch is not yet pushed")
            .addLabeledComponent(remoteUrlReplacementsLabel, remoteUrlReplacementsTextField)
            .addTooltip("You can replace specified strings in your repo's remote URL.")
            .addTooltip(
                "Use any number of pairs: \"search1, replacement1, search2, replacement2, ...\".")
            .addTooltip(
                "Pairs are replaced from left to right. Whitespace around commas doesn't matter.")
            .addComponent(isUrlNotificationDismissedCheckBox, 10)
            .getPanel();
    navigationSettingsPanel.setBorder(
        IdeBorderFactory.createTitledBorder("Navigation Settings", true, JBUI.insetsTop(8)));
    return navigationSettingsPanel;
  }

  @NotNull
  private JPanel createCodySettingsPanel() {
    isCodyAutoCompleteEnabledCheckBox = new JBCheckBox("Enable Cody autocomplete");
    JPanel codySettingsPanel =
        FormBuilder.createFormBuilder()
            .addComponent(isCodyAutoCompleteEnabledCheckBox, 10)
            .getPanel();
    codySettingsPanel.setBorder(
        IdeBorderFactory.createTitledBorder("Cody Settings", true, JBUI.insetsTop(8)));
    return codySettingsPanel;
  }
}
