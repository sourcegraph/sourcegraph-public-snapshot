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
public
class SourcegraphConfig implements PersistentStateComponent<SourcegraphConfig> {
    public String url;
    public String defaultBranch;
    public String remoteUrlReplacements;
    public String lastSearchQuery;
    public boolean lastSearchCaseSensitive;
    public String lastSearchPatternType;
    public String lastSearchContextSpec;
    public boolean isGlobbingEnabled;
    public String accessToken;

    @Nullable
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

    @NotNull
    public Search getLastSearch() {
        return new Search(lastSearchQuery, lastSearchCaseSensitive, lastSearchPatternType, lastSearchContextSpec);
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
    public void loadState(@NotNull SourcegraphConfig config) {
        this.url = config.url;
        this.defaultBranch = config.defaultBranch;
        this.remoteUrlReplacements = config.remoteUrlReplacements;
        this.lastSearchQuery = config.lastSearchQuery != null ? config.lastSearchQuery : "";
        this.lastSearchCaseSensitive = config.lastSearchCaseSensitive;
        this.lastSearchPatternType = config.lastSearchPatternType != null ? config.lastSearchPatternType : "literal";
        this.lastSearchContextSpec = config.lastSearchContextSpec != null ? config.lastSearchContextSpec : "global";
        this.isGlobbingEnabled = config.isGlobbingEnabled;
        this.accessToken = config.accessToken;
    }
}
