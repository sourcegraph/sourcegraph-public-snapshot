package com.sourcegraph.config;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

@State(
    name = "ApplicationConfig",
    storages = {@Storage("sourcegraph.xml")})
public class SourcegraphApplicationService implements PersistentStateComponent<SourcegraphApplicationService> {
    @Nullable
    public String instanceType;
    @Nullable
    public String url;
    @Nullable
    public String accessToken;
    @Nullable
    public String customRequestHeaders;
    @Nullable
    public String defaultBranch;
    @Nullable
    public String remoteUrlReplacements;
    public boolean isGlobbingEnabled; // This can be a primitive boolean, we need no "null" state
    @Nullable
    public String anonymousUserId;
    public boolean isInstallEventLogged;
    public boolean isUrlNotificationDismissed;
    @Nullable
    public Boolean authenticationFailedLastTime;
    @Nullable
    public String lastUpdateNotificationPluginVersion; // The version of the plugin that last notified the user about an update

    @NotNull
    public static SourcegraphApplicationService getInstance() {
        return ApplicationManager.getApplication()
            .getService(SourcegraphApplicationService.class);
    }

    @Nullable
    public String getInstanceType() {
        return instanceType;
    }

    @Nullable
    public String getSourcegraphUrl() {
        return url;
    }

    @Nullable
    public String getAccessToken() {
        return accessToken;
    }

    @Nullable
    public String getCustomRequestHeaders() {
        return customRequestHeaders;
    }

    @Nullable
    public String getDefaultBranchName() {
        return defaultBranch;
    }

    @Nullable
    public String getRemoteUrlReplacements() {
        return remoteUrlReplacements;
    }

    public boolean isGlobbingEnabled() {
        return this.isGlobbingEnabled;
    }

    @Nullable
    public String getAnonymousUserId() {
        return anonymousUserId;
    }

    public boolean isInstallEventLogged() {
        return isInstallEventLogged;
    }

    public boolean isUrlNotificationDismissed() {
        return isUrlNotificationDismissed;
    }

    @Nullable
    public Boolean getAuthenticationFailedLastTime() {
        return authenticationFailedLastTime;
    }

    @Nullable
    public String getLastUpdateNotificationPluginVersion() {
        return lastUpdateNotificationPluginVersion;
    }

    @Nullable
    @Override
    public SourcegraphApplicationService getState() {
        return this;
    }

    @Override
    public void loadState(@NotNull SourcegraphApplicationService settings) {
        this.instanceType = settings.instanceType;
        this.url = settings.url;
        this.accessToken = settings.accessToken;
        this.customRequestHeaders = settings.customRequestHeaders;
        this.defaultBranch = settings.defaultBranch;
        this.remoteUrlReplacements = settings.remoteUrlReplacements;
        this.isGlobbingEnabled = settings.isGlobbingEnabled;
        this.anonymousUserId = settings.anonymousUserId;
        this.isUrlNotificationDismissed = settings.isUrlNotificationDismissed;
        this.authenticationFailedLastTime = settings.authenticationFailedLastTime;
        this.lastUpdateNotificationPluginVersion = settings.lastUpdateNotificationPluginVersion;
    }
}
