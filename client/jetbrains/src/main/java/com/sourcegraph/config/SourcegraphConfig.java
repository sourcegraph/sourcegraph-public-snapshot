package com.sourcegraph.config;

import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import com.intellij.openapi.project.Project;
import com.sourcegraph.find.Search;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

@State(
    name = "Config",
    storages = {@Storage("sourcegraph.xml")})
public class SourcegraphConfig implements PersistentStateComponent<SourcegraphConfig> {
    public String url;
    public String defaultBranch;
    public String remoteUrlReplacements;
    public String lastSearchQuery;
    public boolean lastSearchCaseSensitive;
    public String lastSearchPatternType;
    public String lastSearchContextSpec;
    public boolean isGlobbingEnabled;
    public String accessToken;

    @NotNull
    public static SourcegraphConfig getInstance(@NotNull Project project) {
        return project.getService(SourcegraphConfig.class);
    }

    @Nullable
    public String getSourcegraphUrl() {
        return url;
    }

    @Nullable
    public String getDefaultBranchName() {
        return defaultBranch;
    }

    @Nullable
    public String getRemoteUrlReplacements() {
        return remoteUrlReplacements;
    }

    @Nullable
    public Search getLastSearch() {
        if (lastSearchQuery == null) {
            return null;
        } else {
            return new Search(lastSearchQuery, lastSearchCaseSensitive, lastSearchPatternType, lastSearchContextSpec);
        }
    }

    public boolean isGlobbingEnabled() {
        return this.isGlobbingEnabled;
    }

    @Nullable
    public String getAccessToken() {
        return accessToken;
    }

    @Nullable
    @Override
    public SourcegraphConfig getState() {
        return this;
    }

    @Override
    public void loadState(@NotNull SourcegraphConfig settings) {
        this.url = settings.url;
        this.defaultBranch = settings.defaultBranch;
        this.remoteUrlReplacements = settings.remoteUrlReplacements;
        this.lastSearchQuery = settings.lastSearchQuery != null ? settings.lastSearchQuery : "";
        this.lastSearchCaseSensitive = settings.lastSearchCaseSensitive;
        this.lastSearchPatternType = settings.lastSearchPatternType != null ? settings.lastSearchPatternType : "literal";
        this.lastSearchContextSpec = settings.lastSearchContextSpec != null ? settings.lastSearchContextSpec : "global";
        this.isGlobbingEnabled = settings.isGlobbingEnabled;
        this.accessToken = settings.accessToken;
    }
}
