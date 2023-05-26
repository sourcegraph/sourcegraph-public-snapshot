package com.sourcegraph.cody.config;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.project.Project;
import com.intellij.util.messages.MessageBus;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/** Provides controller functionality for application settings. */
public class SettingsConfigurableHelper {
  public static boolean isModified(
      @Nullable Project project, @NotNull SettingsComponent newSettings) {
    CodyService oldSettings =
        project != null
            ? CodyProjectService.getInstance(project)
            : CodyApplicationService.getInstance();
    return !newSettings
            .getInstanceType()
            .toString()
            .equals(
                oldSettings.getInstanceType() != null
                    ? oldSettings.getInstanceType()
                    : SettingsComponent.InstanceType.DOTCOM.toString())
        || !(newSettings.getDotcomAccessToken().equals(oldSettings.getDotcomAccessToken())
            || newSettings.getDotcomAccessToken().isEmpty()
                && oldSettings.getDotcomAccessToken() == null)
        || !newSettings
            .getEnterpriseUrl()
            .equals(oldSettings.getEnterpriseUrl() != null ? oldSettings.getEnterpriseUrl() : "")
        || !(newSettings.getEnterpriseAccessToken().equals(oldSettings.getEnterpriseAccessToken())
            || newSettings.getEnterpriseAccessToken().isEmpty()
                && oldSettings.getEnterpriseAccessToken() == null)
        || !newSettings
            .getCustomRequestHeaders()
            .equals(
                oldSettings.getCustomRequestHeaders() != null
                    ? oldSettings.getCustomRequestHeaders()
                    : "")
        || !newSettings
            .getCodebase()
            .equals(oldSettings.getCodebase() != null ? oldSettings.getCodebase() : "")
        || newSettings.areChatPredictionsEnabled()
            != Boolean.TRUE.equals(oldSettings.areChatPredictionsEnabled());
  }

  public static void apply(@Nullable Project project, @NotNull SettingsComponent settings) {
    // Get message bus and publisher
    MessageBus bus =
        project != null
            ? project.getMessageBus()
            : ApplicationManager.getApplication().getMessageBus();
    PluginSettingChangeActionNotifier publisher =
        bus.syncPublisher(PluginSettingChangeActionNotifier.TOPIC);

    // Select settings service: application or project
    CodyService apSettings =
        project != null
            ? CodyProjectService.getInstance(project)
            : CodyApplicationService.getInstance();

    // Get old and new settings
    String oldDotcomAccessToken = ConfigUtil.getDotcomAccessToken(project);
    String oldEnterpriseUrl = ConfigUtil.getSourcegraphUrl(project);
    String oldEnterpriseAccessToken = ConfigUtil.getEnterpriseAccessToken(project);
    String newDotcomAccessToken = settings.getDotcomAccessToken();
    String newEnterpriseUrl = settings.getEnterpriseUrl();
    String newEnterpriseAccessToken = settings.getEnterpriseAccessToken();
    String newCustomRequestHeaders = settings.getCustomRequestHeaders();

    // Create context
    PluginSettingChangeContext context =
        new PluginSettingChangeContext(
            oldDotcomAccessToken,
            oldEnterpriseUrl,
            oldEnterpriseAccessToken,
            newEnterpriseUrl,
            newDotcomAccessToken,
            newEnterpriseAccessToken,
            newCustomRequestHeaders);

    // Notify listeners
    publisher.beforeAction(context);

    // Update settings
    String instanceTypeName = settings.getInstanceType().name();
    apSettings.setInstanceType(
        instanceTypeName.equals(SettingsComponent.InstanceType.ENTERPRISE.name())
                || !newDotcomAccessToken.equals("")
            ? instanceTypeName
            : null);
    apSettings.setDotcomAccessToken(!newDotcomAccessToken.equals("") ? newDotcomAccessToken : null);
    apSettings.setEnterpriseUrl(!newEnterpriseUrl.equals("") ? newEnterpriseUrl : null);
    apSettings.setEnterpriseAccessToken(
        !newEnterpriseAccessToken.equals("") ? newEnterpriseAccessToken : null);
    apSettings.setCustomRequestHeaders(settings.getCustomRequestHeaders());
    apSettings.setCodebase(!settings.getCodebase().equals("") ? settings.getCodebase() : null);
    apSettings.setChatPredictionsEnabled(settings.areChatPredictionsEnabled());

    // Notify listeners
    publisher.afterAction(context);
  }

  public static void reset(
      @Nullable Project project, @NotNull SettingsComponent mySettingsComponent) {
    CodyService settings =
        project != null
            ? CodyProjectService.getInstance(project)
            : CodyApplicationService.getInstance();
    String instanceType = settings.getInstanceType();
    mySettingsComponent.setInstanceType(
        instanceType != null
            ? instanceType.equals(SettingsComponent.InstanceType.ENTERPRISE.name())
                ? SettingsComponent.InstanceType.ENTERPRISE
                : SettingsComponent.InstanceType.DOTCOM
            : SettingsComponent.InstanceType.DOTCOM);
    String dotcomAccessToken = settings.getDotcomAccessToken();
    mySettingsComponent.setDotcomAccessToken(dotcomAccessToken != null ? dotcomAccessToken : "");
    mySettingsComponent.setEnterpriseUrl(settings.getEnterpriseUrl());
    String enterpriseAccessToken = settings.getEnterpriseAccessToken();
    mySettingsComponent.setEnterpriseAccessToken(
        enterpriseAccessToken != null ? enterpriseAccessToken : "");
    mySettingsComponent.setCustomRequestHeaders(
        settings.getCustomRequestHeaders() != null ? settings.getCustomRequestHeaders() : "");
    String codebase = settings.getCodebase();
    mySettingsComponent.setCodebase(codebase != null ? codebase : "");
    mySettingsComponent.setAreChatPredictionsEnabled(
        settings.areChatPredictionsEnabled() != null
            && Boolean.TRUE.equals(settings.areChatPredictionsEnabled()));
  }
}
