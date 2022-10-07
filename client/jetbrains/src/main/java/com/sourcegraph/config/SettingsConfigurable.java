package com.sourcegraph.config;

import com.intellij.openapi.options.Configurable;
import com.intellij.openapi.project.Project;
import com.intellij.util.messages.MessageBus;
import org.jetbrains.annotations.Nls;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;

/**
 * Provides controller functionality for application settings.
 */
public class SettingsConfigurable implements Configurable {
    private final Project project;
    private SettingsComponent mySettingsComponent;

    public SettingsConfigurable(Project project) {
        this.project = project;
    }

    @Nls(capitalization = Nls.Capitalization.Title)
    @Override
    public String getDisplayName() {
        return "Sourcegraph";
    }

    @Override
    public JComponent getPreferredFocusedComponent() {
        return mySettingsComponent.getPreferredFocusedComponent();
    }

    @Nullable
    @Override
    public JComponent createComponent() {
        mySettingsComponent = new SettingsComponent(project);
        return mySettingsComponent.getPanel();
    }

    @Override
    public boolean isModified() {
        return !mySettingsComponent.getInstanceType().equals(ConfigUtil.getInstanceType(project))
            || !mySettingsComponent.getEnterpriseUrl().equals(ConfigUtil.getEnterpriseUrl(project))
            || !(mySettingsComponent.getAccessToken().equals(ConfigUtil.getAccessToken(project)) || mySettingsComponent.getAccessToken().isEmpty() && ConfigUtil.getAccessToken(project) == null)
            || !mySettingsComponent.getCustomRequestHeaders().equals(ConfigUtil.getCustomRequestHeaders(project))
            || !mySettingsComponent.getDefaultBranchName().equals(ConfigUtil.getDefaultBranchName(project))
            || !mySettingsComponent.getRemoteUrlReplacements().equals(ConfigUtil.getRemoteUrlReplacements(project))
            || mySettingsComponent.isGlobbingEnabled() != ConfigUtil.isGlobbingEnabled(project)
            || mySettingsComponent.isUrlNotificationDismissed() != ConfigUtil.isUrlNotificationDismissed();
    }

    @Override
    public void apply() {
        MessageBus bus = project.getMessageBus();
        PluginSettingChangeActionNotifier publisher = bus.syncPublisher(PluginSettingChangeActionNotifier.TOPIC);

        SourcegraphApplicationService aSettings = SourcegraphApplicationService.getInstance();
        SourcegraphProjectService pSettings = SourcegraphProjectService.getInstance(project);

        String oldUrl = ConfigUtil.getSourcegraphUrl(project);
        String oldAccessToken = ConfigUtil.getAccessToken(project);
        String newUrl = mySettingsComponent.getEnterpriseUrl();
        String newAccessToken = mySettingsComponent.getAccessToken();
        String newCustomRequestHeaders = mySettingsComponent.getCustomRequestHeaders();
        PluginSettingChangeContext context = new PluginSettingChangeContext(oldUrl, oldAccessToken, newUrl, newAccessToken, newCustomRequestHeaders);

        publisher.beforeAction(context);

        if (pSettings.instanceType != null) {
            pSettings.instanceType = mySettingsComponent.getInstanceType().name();
        } else {
            aSettings.instanceType = mySettingsComponent.getInstanceType().name();
        }
        if (pSettings.url != null) {
            pSettings.url = newUrl;
        } else {
            aSettings.url = newUrl;
        }
        if (pSettings.accessToken != null) {
            pSettings.accessToken = newAccessToken;
        } else {
            aSettings.accessToken = newAccessToken;
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
        //noinspection ReplaceNullCheck
        if (pSettings.isGlobbingEnabled != null) {
            pSettings.isGlobbingEnabled = mySettingsComponent.isGlobbingEnabled();
        } else {
            aSettings.isGlobbingEnabled = mySettingsComponent.isGlobbingEnabled();
        }
        aSettings.isUrlNotificationDismissed = mySettingsComponent.isUrlNotificationDismissed();

        publisher.afterAction(context);
    }

    @Override
    public void reset() {
        mySettingsComponent.setInstanceType(ConfigUtil.getInstanceType(project));
        mySettingsComponent.setEnterpriseUrl(ConfigUtil.getEnterpriseUrl(project));
        String accessToken = ConfigUtil.getAccessToken(project);
        mySettingsComponent.setAccessToken(accessToken != null ? accessToken : "");
        mySettingsComponent.setCustomRequestHeaders(ConfigUtil.getCustomRequestHeaders(project));
        String defaultBranchName = ConfigUtil.getDefaultBranchName(project);
        mySettingsComponent.setDefaultBranchName(defaultBranchName);
        String remoteUrlReplacements = ConfigUtil.getRemoteUrlReplacements(project);
        mySettingsComponent.setRemoteUrlReplacements(remoteUrlReplacements);
        mySettingsComponent.setGlobbingEnabled(ConfigUtil.isGlobbingEnabled(project));
        mySettingsComponent.setUrlNotificationDismissedEnabled(ConfigUtil.isUrlNotificationDismissed());
    }

    @Override
    public void disposeUIResources() {
        mySettingsComponent = null;
    }

}
