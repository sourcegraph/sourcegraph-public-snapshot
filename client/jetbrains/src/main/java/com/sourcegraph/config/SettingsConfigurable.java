package com.sourcegraph.config;

import com.intellij.openapi.options.Configurable;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.util.messages.MessageBus;
import com.sourcegraph.cody.localapp.LocalAppManager;
import java.util.Objects;
import javax.swing.*;
import org.apache.commons.lang3.StringUtils;
import org.jetbrains.annotations.Nls;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/** Provides controller functionality for application settings. */
public class SettingsConfigurable implements Configurable {
  private final @NotNull Project project;
  private final @NotNull SettingsComponent mySettingsComponent;

  public SettingsConfigurable(@NotNull Project project) {
    this.project = project;
    mySettingsComponent = new SettingsComponent(project);
  }

  @Nls(capitalization = Nls.Capitalization.Title)
  @Override
  public String getDisplayName() {
    return "Sourcegraph Cody + Code Search";
  }

  @Override
  public JComponent getPreferredFocusedComponent() {
    return mySettingsComponent.getPreferredFocusedComponent();
  }

  @Nullable
  @Override
  public JComponent createComponent() {
    Disposer.register(project, mySettingsComponent);
    return mySettingsComponent.getPanel();
  }

  @Override
  public boolean isModified() {
    return !mySettingsComponent.getInstanceType().equals(ConfigUtil.getInstanceType(project))
        || !mySettingsComponent.getEnterpriseUrl().equals(ConfigUtil.getEnterpriseUrl(project))
        || mySettingsComponent.isDotComAccessTokenChanged()
        || mySettingsComponent.isEnterpriseAccessTokenChanged()
        || !mySettingsComponent
            .getCustomRequestHeaders()
            .equals(ConfigUtil.getCustomRequestHeaders(project))
        || !mySettingsComponent
            .getDefaultBranchName()
            .equals(ConfigUtil.getDefaultBranchName(project))
        || !mySettingsComponent
            .getRemoteUrlReplacements()
            .equals(ConfigUtil.getRemoteUrlReplacements(project))
        || mySettingsComponent.isUrlNotificationDismissed()
            != ConfigUtil.isUrlNotificationDismissed()
        || mySettingsComponent.isCodyEnabled() != ConfigUtil.isCodyEnabled()
        || mySettingsComponent.isCodyAutocompleteEnabled() != ConfigUtil.isCodyAutocompleteEnabled()
        || mySettingsComponent.isCodyDebugEnabled() != ConfigUtil.isCodyDebugEnabled()
        || mySettingsComponent.isCodyVerboseDebugEnabled() != ConfigUtil.isCodyVerboseDebugEnabled()
        || mySettingsComponent.isCustomAutocompleteColorEnabled()
            != ConfigUtil.isCustomAutocompleteColorEnabled()
        || !Objects.equals(
            mySettingsComponent.getCustomAutocompleteColorPanel(),
            ConfigUtil.getCustomAutocompleteColor());
  }

  @Override
  public void apply() {
    MessageBus bus = project.getMessageBus();
    PluginSettingChangeActionNotifier publisher =
        bus.syncPublisher(PluginSettingChangeActionNotifier.TOPIC);

    CodyApplicationService aSettings = CodyApplicationService.getInstance();
    CodyProjectService pSettings = CodyProjectService.getInstance(project);

    boolean oldCodyEnabled = ConfigUtil.isCodyEnabled();
    boolean oldCodyAutocompleteEnabled = ConfigUtil.isCodyAutocompleteEnabled();
    boolean oldCodyDebugEnabled = ConfigUtil.isCodyDebugEnabled();
    boolean oldCodyVerboseDebugEnabled = ConfigUtil.isCodyVerboseDebugEnabled();
    boolean oldIsCustomAutocompleteColorEnabled = ConfigUtil.isCustomAutocompleteColorEnabled();
    Integer oldCustomAutocompleteColor = ConfigUtil.getCustomAutocompleteColor();
    String oldUrl = ConfigUtil.getSourcegraphUrl(project);
    String newDotComAccessToken = mySettingsComponent.getDotComAccessToken();
    String newEnterpriseAccessToken = mySettingsComponent.getEnterpriseAccessToken();
    String enterpriseUrl = mySettingsComponent.getEnterpriseUrl();
    SettingsComponent.InstanceType newInstanceType = mySettingsComponent.getInstanceType();
    String newUrl;
    if (newInstanceType.equals(SettingsComponent.InstanceType.DOTCOM)) {
      newUrl = ConfigUtil.DOTCOM_URL;
    } else if (newInstanceType.equals(SettingsComponent.InstanceType.ENTERPRISE)) {
      newUrl = enterpriseUrl;
    } else {
      newUrl = LocalAppManager.getLocalAppUrl();
    }
    PluginSettingChangeContext context =
        new PluginSettingChangeContext(
            oldCodyEnabled,
            oldCodyAutocompleteEnabled,
            oldUrl,
            oldCodyDebugEnabled,
            oldCodyVerboseDebugEnabled,
            oldCustomAutocompleteColor,
            oldIsCustomAutocompleteColorEnabled,
            newUrl,
            mySettingsComponent.isDotComAccessTokenChanged(),
            mySettingsComponent.isEnterpriseAccessTokenChanged(),
            mySettingsComponent.getCustomRequestHeaders(),
            mySettingsComponent.isCodyEnabled(),
            mySettingsComponent.isCodyAutocompleteEnabled(),
            mySettingsComponent.isCodyDebugEnabled(),
            mySettingsComponent.isCodyVerboseDebugEnabled(),
            mySettingsComponent.isCustomAutocompleteColorEnabled(),
            mySettingsComponent.getCustomAutocompleteColorPanel());

    publisher.beforeAction(context);

    if (pSettings.instanceType != null) {
      pSettings.instanceType = newInstanceType.name();
    } else {
      aSettings.instanceType = newInstanceType.name();
    }
    if (pSettings.url != null) {
      pSettings.url = enterpriseUrl;
    } else {
      aSettings.url = enterpriseUrl;
    }
    if (newInstanceType == SettingsComponent.InstanceType.DOTCOM && newDotComAccessToken != null) {
      if (pSettings.dotComAccessToken != null) {
        pSettings.dotComAccessToken = newDotComAccessToken;
      } else {
        aSettings.setSafeDotComAccessToken(newDotComAccessToken);
        aSettings.isDotComAccessTokenSet = StringUtils.isNotEmpty(newDotComAccessToken);
      }
    }
    if (newInstanceType == SettingsComponent.InstanceType.ENTERPRISE
        && newEnterpriseAccessToken != null) {
      if (pSettings.enterpriseAccessToken != null) {
        pSettings.enterpriseAccessToken = newEnterpriseAccessToken;
      } else {
        aSettings.setSafeEnterpriseAccessToken(newEnterpriseAccessToken);
        aSettings.isEnterpriseAccessTokenSet = StringUtils.isNotEmpty(newEnterpriseAccessToken);
      }
    }
    if (pSettings.customRequestHeaders != null) {
      pSettings.customRequestHeaders = mySettingsComponent.getCustomRequestHeaders();
    } else {
      aSettings.customRequestHeaders = mySettingsComponent.getCustomRequestHeaders();
    }
    if (pSettings.defaultBranch != null) {
      pSettings.defaultBranch = mySettingsComponent.getDefaultBranchName();
    } else {
      aSettings.defaultBranch = mySettingsComponent.getDefaultBranchName();
    }
    if (pSettings.remoteUrlReplacements != null) {
      pSettings.remoteUrlReplacements = mySettingsComponent.getRemoteUrlReplacements();
    } else {
      aSettings.remoteUrlReplacements = mySettingsComponent.getRemoteUrlReplacements();
    }
    aSettings.isUrlNotificationDismissed = mySettingsComponent.isUrlNotificationDismissed();
    aSettings.setCodyEnabled(mySettingsComponent.isCodyEnabled());
    aSettings.isCodyAutocompleteEnabled = mySettingsComponent.isCodyAutocompleteEnabled();
    aSettings.isCodyDebugEnabled = mySettingsComponent.isCodyDebugEnabled();
    aSettings.isCodyVerboseDebugEnabled = mySettingsComponent.isCodyVerboseDebugEnabled();
    aSettings.isCustomAutocompleteColorEnabled =
        mySettingsComponent.isCustomAutocompleteColorEnabled();
    aSettings.customAutocompleteColor = mySettingsComponent.getCustomAutocompleteColorPanel();

    publisher.afterAction(context);
  }

  @Override
  public void reset() {
    mySettingsComponent.setInstanceType(ConfigUtil.getInstanceType(project));
    mySettingsComponent.setEnterpriseUrl(ConfigUtil.getEnterpriseUrl(project));
    mySettingsComponent.resetDotComAccessToken();
    mySettingsComponent.resetEnterpriseAccessToken();
    mySettingsComponent.setCustomRequestHeaders(ConfigUtil.getCustomRequestHeaders(project));
    String defaultBranchName = ConfigUtil.getDefaultBranchName(project);
    mySettingsComponent.setDefaultBranchName(defaultBranchName);
    String remoteUrlReplacements = ConfigUtil.getRemoteUrlReplacements(project);
    mySettingsComponent.setRemoteUrlReplacements(remoteUrlReplacements);
    mySettingsComponent.setUrlNotificationDismissedEnabled(ConfigUtil.isUrlNotificationDismissed());
    mySettingsComponent.setCodyEnabled(ConfigUtil.isCodyEnabled());
    mySettingsComponent.setCodyAutocompleteEnabled(ConfigUtil.isCodyAutocompleteEnabled());
    mySettingsComponent.setIsCodyDebugEnabled(ConfigUtil.isCodyDebugEnabled());
    mySettingsComponent.setIsCodyVerboseDebugEnabled(ConfigUtil.isCodyVerboseDebugEnabled());
    mySettingsComponent.setIsCustomAutocompleteColorEnabled(
        ConfigUtil.isCustomAutocompleteColorEnabled());
    mySettingsComponent.setCustomAutocompleteColorPanel(ConfigUtil.getCustomAutocompleteColor());
    mySettingsComponent.getPanel().requestFocusInWindow();
  }

  @Override
  public void disposeUIResources() {
    mySettingsComponent.dispose();
  }
}
