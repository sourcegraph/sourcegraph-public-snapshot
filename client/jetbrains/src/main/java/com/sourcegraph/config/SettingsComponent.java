package com.sourcegraph.config;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.application.ModalityState;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.ComponentValidator;
import com.intellij.openapi.ui.ValidationInfo;
import com.intellij.ui.*;
import com.intellij.ui.components.ActionLink;
import com.intellij.ui.components.JBCheckBox;
import com.intellij.ui.components.JBLabel;
import com.intellij.ui.components.JBPasswordField;
import com.intellij.ui.components.JBTextField;
import com.intellij.util.ui.FormBuilder;
import com.intellij.util.ui.JBUI;
import com.intellij.util.ui.UIUtil;
import com.jetbrains.jsonSchema.settings.mappings.JsonSchemaConfigurable;
import com.sourcegraph.cody.localapp.LocalAppManager;
import com.sourcegraph.cody.ui.PasswordFieldWithShowHideButton;
import com.sourcegraph.common.AuthorizationUtil;
import java.awt.*;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.util.Enumeration;
import java.util.Objects;
import java.util.Optional;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;
import java.util.function.Supplier;
import javax.swing.*;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.text.Document;
import javax.swing.text.JTextComponent;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/** Supports creating and managing a {@link JPanel} for the Settings Dialog. */
public class SettingsComponent implements Disposable {
  private final JPanel panel;
  private ButtonGroup instanceTypeButtonGroup;
  private JBTextField urlTextField;
  private PasswordFieldWithShowHideButton enterpriseAccessTokenTextField;
  private PasswordFieldWithShowHideButton dotComAccessTokenTextField;
  private JBLabel userDocsLinkComment;
  private JBLabel enterpriseAccessTokenLinkComment;
  private JBTextField customRequestHeadersTextField;
  private JBTextField defaultBranchNameTextField;
  private JBTextField remoteUrlReplacementsTextField;
  private JBCheckBox isUrlNotificationDismissedCheckBox;
  private JBCheckBox isCodyEnabledCheckBox;
  private JBCheckBox isCodyAutocompleteEnabledCheckBox;

  private ActionLink installLocalAppLink;
  private JLabel installLocalAppComment;
  private ActionLink runLocalAppLink;
  private JLabel runLocalAppComment;
  private JBCheckBox isCodyDebugEnabledCheckBox;
  private JBCheckBox isCodyVerboseDebugEnabledCheckBox;
  private JBCheckBox isCustomAutocompleteColorEnabledCheckBox;
  private ColorPanel customAutocompleteColorPanel;

  private final ScheduledExecutorService codyAppStateCheckerExecutorService =
      Executors.newSingleThreadScheduledExecutor();
  private final int colorPanelWidth = 62;

  public JComponent getPreferredFocusedComponent() {
    return defaultBranchNameTextField;
  }

  public SettingsComponent(@NotNull Project project) {
    JPanel userAuthenticationPanel = createAuthenticationPanel(project);
    JPanel navigationSettingsPanel = createNavigationSettingsPanel();
    JPanel codySettingsPanel = createCodySettingsPanel();

    panel =
        FormBuilder.createFormBuilder()
            .addComponent(userAuthenticationPanel)
            .addComponent(codySettingsPanel)
            .addComponent(navigationSettingsPanel)
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
  private JPanel createAuthenticationPanel(@NotNull Project project) {
    // Create URL field for the enterprise section
    JBLabel urlLabel = new JBLabel("Sourcegraph URL:");
    urlTextField = new JBTextField();
    //noinspection DialogTitleCapitalization
    urlTextField.getEmptyText().setText("https://sourcegraph.example.com");
    urlTextField.setToolTipText("The default is \"" + ConfigUtil.DOTCOM_URL + "\".");
    addValidation(
        urlTextField,
        () ->
            urlTextField.getText().isEmpty()
                ? new ValidationInfo("Missing URL", urlTextField)
                : (!JsonSchemaConfigurable.isValidURL(urlTextField.getText())
                    ? new ValidationInfo("This is an invalid URL", urlTextField)
                    : null));
    addDocumentListener(
        urlTextField, urlTextField.getDocument(), e -> updateAccessTokenLinkCommentText());

    // Create enterprise access token field
    JBLabel accessTokenLabel = new JBLabel("Access token:");
    enterpriseAccessTokenTextField =
        new PasswordFieldWithShowHideButton(
            new JBPasswordField(), () -> ConfigUtil.getEnterpriseAccessToken(project));
    enterpriseAccessTokenTextField.setEmptyText("Paste your access token here");
    addValidation(
        enterpriseAccessTokenTextField,
        () -> {
          String password = enterpriseAccessTokenTextField.getPassword();
          return password != null && !AuthorizationUtil.isValidAccessToken(password)
              ? new ValidationInfo("Invalid access token", enterpriseAccessTokenTextField)
              : null;
        });

    // Create dotcom access token field
    JBLabel dotComAccessTokenComment =
        new JBLabel(
                "(optional) To use Cody, you will need an access token to sign in.",
                UIUtil.ComponentStyle.SMALL,
                UIUtil.FontColor.BRIGHTER)
            .withBorder(JBUI.Borders.emptyLeft(10));
    JBLabel dotComAccessTokenLabel = new JBLabel("Access token:");
    dotComAccessTokenTextField =
        new PasswordFieldWithShowHideButton(
            new JBPasswordField(), () -> ConfigUtil.getDotComAccessToken(project));
    dotComAccessTokenTextField.setEmptyText("Paste your access token here");
    addValidation(
        dotComAccessTokenTextField,
        () -> {
          String password = dotComAccessTokenTextField.getPassword();
          return password != null && !AuthorizationUtil.isValidAccessToken(password)
              ? new ValidationInfo("Invalid access token", dotComAccessTokenTextField)
              : null;
        });

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
    @SuppressWarnings("DialogTitleCapitalization")
    String codyAppCommentText =
        isLocalAppPlatformSupported
            ? "Use Sourcegraph through Cody App."
            : platformName
                + " is not yet supported by the Cody App. Keep an eye on future updates!";
    JBLabel codyAppComment =
        new JBLabel(codyAppCommentText, UIUtil.ComponentStyle.SMALL, UIUtil.FontColor.BRIGHTER);
    codyAppComment.setBorder(JBUI.Borders.emptyLeft(20));
    installLocalAppComment =
        new JBLabel(
            "Cody App wasn't detected on this system, it seems it hasn't been installed yet.",
            UIUtil.ComponentStyle.SMALL,
            UIUtil.FontColor.BRIGHTER);
    installLocalAppComment.setVisible(false);
    installLocalAppComment.setBorder(JBUI.Borders.emptyLeft(20));
    installLocalAppLink =
        simpleActionLink("Install Cody App...", LocalAppManager::browseLocalAppInstallPage);
    installLocalAppLink.setVisible(false);
    installLocalAppLink.setBorder(JBUI.Borders.emptyLeft(20));
    runLocalAppLink = simpleActionLink("Run Cody App...", LocalAppManager::runLocalApp);
    runLocalAppLink.setVisible(false);
    runLocalAppLink.setBorder(JBUI.Borders.emptyLeft(20));
    runLocalAppComment =
        new JBLabel(
            "Cody App seems to be installed, but it's not running, currently.",
            UIUtil.ComponentStyle.SMALL,
            UIUtil.FontColor.BRIGHTER);
    runLocalAppComment.setVisible(false);
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
          if (customRequestHeadersTextField.getText().isEmpty()) {
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
        IdeBorderFactory.createTitledBorder("Authentication", true, JBUI.insetsTop(8)));

    updateVisibilityOfHelperLinks();
    codyAppStateCheckerExecutorService.scheduleWithFixedDelay(
        this::updateVisibilityOfHelperLinks, 0, 1, TimeUnit.SECONDS);
    return userAuthenticationPanel;
  }

  private void updateVisibilityOfHelperLinks() {
    boolean isLocalAppInstalled = LocalAppManager.isLocalAppInstalled();
    boolean isLocalAppRunning = LocalAppManager.isLocalAppRunning();
    boolean isLocalAppPlatformSupported = LocalAppManager.isPlatformSupported();
    boolean shouldShowInstallLocalAppLink = !isLocalAppInstalled && isLocalAppPlatformSupported;
    boolean shouldShowRunLocalAppLink = isLocalAppInstalled && !isLocalAppRunning;
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              installLocalAppComment.setVisible(shouldShowInstallLocalAppLink);
              installLocalAppLink.setVisible(shouldShowInstallLocalAppLink);
              runLocalAppLink.setVisible(shouldShowRunLocalAppLink);
              runLocalAppComment.setVisible(shouldShowRunLocalAppLink);
            },
            ModalityState.any());
  }

  @NotNull
  public String getEnterpriseUrl() {
    return urlTextField.getText();
  }

  public void setEnterpriseUrl(@Nullable String value) {
    urlTextField.setText(value != null ? value : "");
  }

  /**
   * @return Null means we don't know the token because it wasn't loaded from the secure storage. An
   *     empty token means the user has explicitly set it to empty.
   */
  @Nullable
  public String getDotComAccessToken() {
    return dotComAccessTokenTextField.getPassword();
  }

  public void resetDotComAccessToken() {
    dotComAccessTokenTextField.resetUI();
  }

  public boolean isDotComAccessTokenChanged() {
    return dotComAccessTokenTextField.hasPasswordChanged();
  }

  /**
   * @return Null means we don't know the token because it wasn't loaded from the secure storage. An
   *     empty token means the user has explicitly set it to empty.
   */
  @Nullable
  public String getEnterpriseAccessToken() {
    return enterpriseAccessTokenTextField.getPassword();
  }

  public void resetEnterpriseAccessToken() {
    enterpriseAccessTokenTextField.resetUI();
  }

  public boolean isEnterpriseAccessTokenChanged() {
    return enterpriseAccessTokenTextField.hasPasswordChanged();
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

  public boolean isCodyEnabled() {
    return isCodyEnabledCheckBox.isSelected();
  }

  public void setCodyEnabled(boolean value) {
    isCodyEnabledCheckBox.setSelected(value);
    this.onDidCodyEnableSettingChange();
    if (!value) {
      setCodyAutocompleteEnabled(false);
    }
  }

  public boolean isCodyAutocompleteEnabled() {
    return isCodyAutocompleteEnabledCheckBox.isSelected();
  }

  public void setCodyAutocompleteEnabled(boolean value) {
    isCodyAutocompleteEnabledCheckBox.setSelected(value);
  }

  public boolean isCodyDebugEnabled() {
    return isCodyDebugEnabledCheckBox.isSelected();
  }

  public void setIsCodyDebugEnabled(boolean value) {
    isCodyDebugEnabledCheckBox.setSelected(value);
  }

  public boolean isCodyVerboseDebugEnabled() {
    return isCodyVerboseDebugEnabledCheckBox.isSelected();
  }

  public void setIsCodyVerboseDebugEnabled(boolean value) {
    isCodyVerboseDebugEnabledCheckBox.setSelected(value);
  }

  public boolean isCustomAutocompleteColorEnabled() {
    return isCustomAutocompleteColorEnabledCheckBox.isSelected();
  }

  public void setIsCustomAutocompleteColorEnabled(boolean value) {
    isCustomAutocompleteColorEnabledCheckBox.setSelected(value);
  }

  public @NotNull Integer getCustomAutocompleteColorPanel() {
    Color selectedColor = customAutocompleteColorPanel.getSelectedColor();
    return Objects.requireNonNull(selectedColor).getRGB();
  }

  public void setCustomAutocompleteColorPanel(Integer value) {
    Color c = new Color(value);
    customAutocompleteColorPanel.setSelectedColor(c);
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
    new ComponentValidator(this).withValidator(validator).installOn(component);
    addDocumentListener(component, component.getDocument(), ComponentValidator::revalidate);
  }

  private void addValidation(
      @NotNull PasswordFieldWithShowHideButton component,
      @NotNull Supplier<ValidationInfo> validator) {
    new ComponentValidator(this).withValidator(validator).installOn(component);
    addDocumentListener(component, component.getDocument(), ComponentValidator::revalidate);
  }

  private void addDocumentListener(
      @NotNull JComponent component,
      @NotNull Document document,
      @NotNull Consumer<ComponentValidator> function) {
    document.addDocumentListener(
        new DocumentListener() {
          @Override
          public void insertUpdate(DocumentEvent e) {
            ComponentValidator.getInstance(component).ifPresent(function);
          }

          @Override
          public void removeUpdate(DocumentEvent e) {
            ComponentValidator.getInstance(component).ifPresent(function);
          }

          @Override
          public void changedUpdate(DocumentEvent e) {
            ComponentValidator.getInstance(component).ifPresent(function);
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
            (!remoteUrlReplacementsTextField.getText().isEmpty()
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
        IdeBorderFactory.createTitledBorder("Code Search", true, JBUI.insetsTop(8)));
    return navigationSettingsPanel;
  }

  @NotNull
  private JPanel createCodySettingsPanel() {
    //noinspection DialogTitleCapitalization
    isCodyEnabledCheckBox = new JBCheckBox("Enable Cody");
    isCodyAutocompleteEnabledCheckBox = new JBCheckBox("Enable Cody autocomplete");
    isCodyDebugEnabledCheckBox = new JBCheckBox("Enable debug");
    isCodyVerboseDebugEnabledCheckBox = new JBCheckBox("Verbose debug");

    isCustomAutocompleteColorEnabledCheckBox = new JBCheckBox("Enable custom autocomplete color");

    customAutocompleteColorPanel = new ColorPanel();
    customAutocompleteColorPanel.setVisible(false);
    isCustomAutocompleteColorEnabledCheckBox.addChangeListener(
        e -> {
          customAutocompleteColorPanel.setVisible(
              isCustomAutocompleteColorEnabledCheckBox.isSelected());
        });

    JPanel customAutocompleteColorPanelWrapper = new JPanel(new FlowLayout(FlowLayout.LEFT));
    customAutocompleteColorPanelWrapper.setBounds(
        0, 0, colorPanelWidth, customAutocompleteColorPanel.getHeight());
    customAutocompleteColorPanelWrapper.add(customAutocompleteColorPanel);

    JPanel codySettingsPanel =
        FormBuilder.createFormBuilder()
            .addComponent(isCodyEnabledCheckBox, 10)
            .addTooltip(
                "Disable this to turn off all AI-based functionality of the plugin, including the Cody chat sidebar and autocomplete")
            .addComponent(isCodyAutocompleteEnabledCheckBox, 5)
            .addComponent(isCodyDebugEnabledCheckBox)
            .addTooltip("Enables debug output visible in the idea.log")
            .addComponent(isCodyVerboseDebugEnabledCheckBox)
            .addLabeledComponent(
                isCustomAutocompleteColorEnabledCheckBox, customAutocompleteColorPanelWrapper)
            .getPanel();
    codySettingsPanel.setBorder(
        IdeBorderFactory.createTitledBorder("Cody AI", true, JBUI.insetsTop(8)));

    // Disable isCodyAutocompleteEnabledCheckBox if isCodyEnabledCheckBox is not selected
    isCodyEnabledCheckBox.addActionListener(e -> this.onDidCodyEnableSettingChange());
    this.onDidCodyEnableSettingChange();

    return codySettingsPanel;
  }

  private void onDidCodyEnableSettingChange() {
    isCodyAutocompleteEnabledCheckBox.setEnabled(isCodyEnabledCheckBox.isSelected());
    isCodyDebugEnabledCheckBox.setEnabled(isCodyEnabledCheckBox.isSelected());
    isCodyVerboseDebugEnabledCheckBox.setEnabled(isCodyEnabledCheckBox.isSelected());
    isCustomAutocompleteColorEnabledCheckBox.setEnabled(isCodyEnabledCheckBox.isSelected());
    customAutocompleteColorPanel.setEnabled(isCodyEnabledCheckBox.isSelected());
  }

  @Override
  public void dispose() {}
}
