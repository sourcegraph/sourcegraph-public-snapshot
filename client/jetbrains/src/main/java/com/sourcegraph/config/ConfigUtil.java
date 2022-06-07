package com.sourcegraph.config;

import com.intellij.openapi.project.Project;
import com.sourcegraph.find.Search;
import org.jetbrains.annotations.Contract;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.Objects;

public class ConfigUtil {
    @Nullable
    public static String getDefaultBranchName(@NotNull Project project) {
        String defaultBranch = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getDefaultBranchName();
        if (defaultBranch == null || defaultBranch.length() == 0) {
            return UserLevelConfig.getDefaultBranchName();
        }
        return defaultBranch;
    }

    @Nullable
    public static String getRemoteUrlReplacements(@NotNull Project project) {
        String replacements = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getRemoteUrlReplacements();
        if (replacements == null || replacements.length() == 0) {
            return UserLevelConfig.getRemoteUrlReplacements();
        }
        return replacements;
    }

    @NotNull
    public static String getSourcegraphUrl(@NotNull Project project) {
        String url = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getSourcegraphUrl();
        if (url == null || url.length() == 0) {
            return UserLevelConfig.getSourcegraphUrl();
        }
        return url.endsWith("/") ? url : url + "/";
    }

    @Nullable
    public static Search getLastSearch(@NotNull Project project) {
        return getProjectLevelConfig(project).getLastSearch();
    }

    public static void setLastSearch(@NotNull Project project, @NotNull Search lastSearch) {
        SourcegraphConfig settings = getProjectLevelConfig(project);
        settings.lastSearchQuery = lastSearch.getQuery() != null ? lastSearch.getQuery() : "";
        settings.lastSearchCaseSensitive = lastSearch.isCaseSensitive();
        settings.lastSearchPatternType = lastSearch.getPatternType() != null ? lastSearch.getPatternType() : "literal";
        settings.lastSearchContextSpec = lastSearch.getSelectedSearchContextSpec() != null ? lastSearch.getSelectedSearchContextSpec() : "global";
    }

    public static boolean isGlobbingEnabled(@NotNull Project project) {
        return getProjectLevelConfig(project).isGlobbingEnabled();
    }

    @Nullable
    public static String getAccessToken(Project project) {
        return getProjectLevelConfig(project).getAccessToken();
    }

    @NotNull
    @Contract(pure = true)
    public static String getVersion() {
        return "v1.2.2";
    }

    @NotNull
    private static SourcegraphConfig getProjectLevelConfig(@NotNull Project project) {
        return Objects.requireNonNull(SourcegraphConfig.getInstance(project));
    }
}
