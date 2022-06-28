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
    public String anonymousUserId;
    public boolean isInstallEventLogged;

    @NotNull
    public static SourcegraphApplicationService getInstance() {
        return ApplicationManager.getApplication()
            .getService(SourcegraphApplicationService.class);
    }

    @Nullable
    public String getAnonymousUserId() {
        return anonymousUserId;
    }

    public boolean isInstallEventLogged() {
        return isInstallEventLogged;
    }

    @Nullable
    @Override
    public SourcegraphApplicationService getState() {
        return this;
    }

    @Override
    public void loadState(@NotNull SourcegraphApplicationService config) {
        this.anonymousUserId = config.anonymousUserId;
    }
}
