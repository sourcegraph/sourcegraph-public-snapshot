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
        mySettingsComponent = new SettingsComponent();
        return mySettingsComponent.getPanel();
    }

    @Override
    public boolean isModified() {
        SourcegraphProjectService settings = SourcegraphProjectService.getInstance(project);
        return !mySettingsComponent.getSourcegraphUrl().equals(settings.getSourcegraphUrl())
            || !mySettingsComponent.getAccessToken().equals(settings.getAccessToken())
            || !mySettingsComponent.getDefaultBranchName().equals(settings.getDefaultBranchName())
            || !mySettingsComponent.getRemoteUrlReplacements().equals(settings.getRemoteUrlReplacements())
            || mySettingsComponent.isGlobbingEnabled() != settings.isGlobbingEnabled();
    }

    @Override
    public void apply() {
        MessageBus bus = project.getMessageBus();
        PluginSettingChangeActionNotifier publisher = bus.syncPublisher(PluginSettingChangeActionNotifier.TOPIC);

        SourcegraphProjectService settings = SourcegraphProjectService.getInstance(project);

        String oldUrl = settings.url;
        String oldAccessToken = settings.accessToken;
        String newUrl = mySettingsComponent.getSourcegraphUrl();
        String newAccessToken = mySettingsComponent.getAccessToken();
        PluginSettingChangeContext context = new PluginSettingChangeContext(oldUrl, oldAccessToken, newUrl, newAccessToken);

        publisher.beforeAction(context);

        settings.url = newUrl;
        settings.accessToken = newAccessToken;
        settings.defaultBranch = mySettingsComponent.getDefaultBranchName();
        settings.remoteUrlReplacements = mySettingsComponent.getRemoteUrlReplacements();
        settings.isGlobbingEnabled = mySettingsComponent.isGlobbingEnabled();

        publisher.afterAction(context);
    }

    @Override
    public void reset() {
        SourcegraphProjectService settings = SourcegraphProjectService.getInstance(project);
        mySettingsComponent.setSourcegraphUrl(settings.getSourcegraphUrl() != null ? settings.getSourcegraphUrl() : "https://sourcegraph.com");
        mySettingsComponent.setAccessToken(settings.getAccessToken() != null ? settings.getAccessToken() : "");
        mySettingsComponent.setDefaultBranchName(settings.getDefaultBranchName() != null ? settings.getDefaultBranchName() : "main");
        mySettingsComponent.setRemoteUrlReplacements(settings.getRemoteUrlReplacements() != null ? settings.getRemoteUrlReplacements() : "");
        mySettingsComponent.setGlobbingEnabled(settings.isGlobbingEnabled());
    }

    @Override
    public void disposeUIResources() {
        mySettingsComponent = null;
    }

}
