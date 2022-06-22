package com.sourcegraph.config;

import com.google.gson.JsonObject;
import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginManagerCore;
import com.intellij.openapi.extensions.PluginId;
import com.intellij.openapi.project.Project;
import com.sourcegraph.find.Search;
import org.jetbrains.annotations.Contract;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.Objects;

public class ConfigUtil {
    @NotNull
    public static JsonObject getConfigAsJson(@NotNull Project project) {
        JsonObject configAsJson = new JsonObject();
        configAsJson.addProperty("instanceURL", ConfigUtil.getSourcegraphUrl(project));
        configAsJson.addProperty("isGlobbingEnabled", ConfigUtil.isGlobbingEnabled(project));
        configAsJson.addProperty("accessToken", ConfigUtil.getAccessToken(project));
        return configAsJson;
    }

    @Nullable
    public static String getDefaultBranchName(@NotNull Project project) {
        String defaultBranch = Objects.requireNonNull(SourcegraphProjectService.getInstance(project)).getDefaultBranchName();
        if (defaultBranch == null || defaultBranch.length() == 0) {
            return UserLevelConfig.getDefaultBranchName();
        }
        return defaultBranch;
    }

    @Nullable
    public static String getRemoteUrlReplacements(@NotNull Project project) {
        String replacements = Objects.requireNonNull(SourcegraphProjectService.getInstance(project)).getRemoteUrlReplacements();
        if (replacements == null || replacements.length() == 0) {
            return UserLevelConfig.getRemoteUrlReplacements();
        }
        return replacements;
    }

    @NotNull
    public static String getSourcegraphUrl(@NotNull Project project) {
        String url = Objects.requireNonNull(SourcegraphProjectService.getInstance(project)).getSourcegraphUrl();
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
        SourcegraphProjectService settings = getProjectLevelConfig(project);
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
    public static String getPluginVersion() {
        IdeaPluginDescriptor plugin = PluginManagerCore.getPlugin(PluginId.getId("com.sourcegraph.jetbrains"));
        return plugin != null ? plugin.getVersion() : "unknown";
    }

    @NotNull
    private static SourcegraphProjectService getProjectLevelConfig(@NotNull Project project) {
        return Objects.requireNonNull(SourcegraphProjectService.getInstance(project));
    }
}
