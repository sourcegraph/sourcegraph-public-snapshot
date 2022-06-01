package com.sourcegraph.config;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import com.intellij.util.xmlb.XmlSerializerUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

/**
 * Supports storing the application settings in a persistent way.
 * The {@link State} and {@link Storage} annotations define the name of the data and the file name where
 * these persistent application settings are stored.
 */
@State(
    name = "com.sourcegraph.config.SourcegraphSettingsState",
    storages = @Storage("SdkSettingsPlugin.xml")
)
public class SourcegraphSettingsState implements PersistentStateComponent<SourcegraphSettingsState> {

    public String userId = "John Q. Public";
    public boolean ideaStatus = false;

    public static SourcegraphSettingsState getInstance() {
        return ApplicationManager.getApplication().getService(SourcegraphSettingsState.class);
    }

    @Nullable
    @Override
    public SourcegraphSettingsState getState() {
        return this;
    }

    @Override
    public void loadState(@NotNull SourcegraphSettingsState state) {
        XmlSerializerUtil.copyBean(state, this);
    }
}
