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

    @Nullable
    public static SourcegraphConfig getInstance(@NotNull Project project) {
        return project.getService(SourcegraphConfig.class);
    }

    public String getSourcegraphUrl() {
        return url;
    }

    public String getDefaultBranchName() {
        return defaultBranch;
    }

    public String getRemoteUrlReplacements() {
        return remoteUrlReplacements;
    }

    public Search getLastSearch() {
        return new Search(lastSearchQuery, lastSearchCaseSensitive, lastSearchPatternType, lastSearchContextSpec);
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
        this.lastSearchQuery = config.lastSearchQuery;
        this.lastSearchCaseSensitive = config.lastSearchCaseSensitive;
        this.lastSearchPatternType = config.lastSearchPatternType;
        this.lastSearchContextSpec = config.lastSearchContextSpec;
    }
}
