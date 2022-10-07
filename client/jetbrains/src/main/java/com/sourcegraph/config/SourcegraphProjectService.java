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
public class SourcegraphProjectService implements PersistentStateComponent<SourcegraphProjectService> {
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
    @Nullable
    public Boolean isGlobbingEnabled; // This needs to be a Boolean rather than a primitive: we need the "null" state
    @Nullable
    public String lastSearchQuery;
    public boolean lastSearchCaseSensitive;
    @Nullable
    public String lastSearchPatternType;
    @Nullable
    public String lastSearchContextSpec;

    @NotNull
    public static SourcegraphProjectService getInstance(@NotNull Project project) {
        return project.getService(SourcegraphProjectService.class);
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

    @Nullable
    public Boolean isGlobbingEnabled() {
        return this.isGlobbingEnabled;
    }

    @Nullable
    public Search getLastSearch() {
        if (lastSearchQuery == null) {
            return null;
        } else {
            return new Search(lastSearchQuery, lastSearchCaseSensitive, lastSearchPatternType, lastSearchContextSpec);
        }
    }

    @Nullable
    @Override
    public SourcegraphProjectService getState() {
        return this;
    }

    @Override
    public void loadState(@NotNull SourcegraphProjectService settings) {
        this.instanceType = settings.instanceType;
        this.url = settings.url;
        this.accessToken = settings.accessToken;
        this.customRequestHeaders = settings.customRequestHeaders;
        this.defaultBranch = settings.defaultBranch;
        this.remoteUrlReplacements = settings.remoteUrlReplacements;
        this.isGlobbingEnabled = settings.isGlobbingEnabled;
        this.lastSearchQuery = settings.lastSearchQuery != null ? settings.lastSearchQuery : "";
        this.lastSearchCaseSensitive = settings.lastSearchCaseSensitive;
        this.lastSearchPatternType = settings.lastSearchPatternType != null ? settings.lastSearchPatternType : "literal";
        this.lastSearchContextSpec = settings.lastSearchContextSpec != null ? settings.lastSearchContextSpec : "global";
    }
}
