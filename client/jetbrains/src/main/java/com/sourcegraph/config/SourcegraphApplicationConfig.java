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
public class SourcegraphApplicationConfig implements PersistentStateComponent<SourcegraphApplicationConfig> {
    @Nullable
    public String anonymousUserId;

    @NotNull
    public static SourcegraphApplicationConfig getInstance() {
        return ApplicationManager.getApplication()
            .getService(SourcegraphApplicationConfig.class);
    }

    @Nullable
    public String getAnonymousUserId() {
        return anonymousUserId;
    }

    @Nullable
    @Override
    public SourcegraphApplicationConfig getState() {
        return this;
    }

    @Override
    public void loadState(@NotNull SourcegraphApplicationConfig config) {
        this.anonymousUserId = config.anonymousUserId;
    }
}
