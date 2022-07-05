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
        configAsJson.addProperty("accessToken", ConfigUtil.getAccessToken(project));
        configAsJson.addProperty("isGlobbingEnabled", ConfigUtil.isGlobbingEnabled(project));
        configAsJson.addProperty("pluginVersion", ConfigUtil.getPluginVersion());
        configAsJson.addProperty("anonymousUserId", ConfigUtil.getAnonymousUserId());
        return configAsJson;
    }

    @NotNull
    public static String getSourcegraphUrl(@NotNull Project project) {
        String projectLevelUrl = Objects.requireNonNull(SourcegraphProjectService.getInstance(project)).getSourcegraphUrl();
        if (projectLevelUrl != null && projectLevelUrl.length() > 0) {
            return addSlashIfNeeded(projectLevelUrl);
        }

        String applicationLevelUrl = Objects.requireNonNull(SourcegraphProjectService.getInstance(project)).getSourcegraphUrl();
        if (applicationLevelUrl != null && applicationLevelUrl.length() > 0) {
            return addSlashIfNeeded(applicationLevelUrl);
        }

        return addSlashIfNeeded(UserLevelConfig.getSourcegraphUrl());
    }

    @NotNull
    private static String addSlashIfNeeded(@NotNull String url) {
        return url.endsWith("/") ? url : url + "/";
    }

    @Nullable
    public static String getAccessToken(Project project) {
        String projectLevelAccessToken = getProjectLevelConfig(project).getAccessToken();
        return projectLevelAccessToken != null ? projectLevelAccessToken : getApplicationLevelConfig().getAccessToken();
    }

    @NotNull
    public static String getDefaultBranchName(@NotNull Project project) {
        String projectLevelDefaultBranchName = Objects.requireNonNull(SourcegraphProjectService.getInstance(project)).getDefaultBranchName();
        if (projectLevelDefaultBranchName != null && projectLevelDefaultBranchName.length() > 0) {
            return projectLevelDefaultBranchName;
        }

        String applicationLevelDefaultBranchName = Objects.requireNonNull(SourcegraphApplicationService.getInstance()).getDefaultBranchName();
        if (applicationLevelDefaultBranchName != null && applicationLevelDefaultBranchName.length() > 0) {
            return applicationLevelDefaultBranchName;
        }

        String userLevelDefaultBranchName = UserLevelConfig.getDefaultBranchName();
        return userLevelDefaultBranchName != null ? userLevelDefaultBranchName : "main";
    }

    @NotNull
    public static String getRemoteUrlReplacements(@NotNull Project project) {
        String projectLevelReplacements = Objects.requireNonNull(SourcegraphProjectService.getInstance(project)).getRemoteUrlReplacements();
        if (projectLevelReplacements != null && projectLevelReplacements.length() > 0) {
            return projectLevelReplacements;
        }

        String applicationLevelReplacements = Objects.requireNonNull(SourcegraphApplicationService.getInstance()).getRemoteUrlReplacements();
        if (applicationLevelReplacements != null && applicationLevelReplacements.length() > 0) {
            return applicationLevelReplacements;
        }

        String userLevelRemoteUrlReplacements = UserLevelConfig.getRemoteUrlReplacements();
        return userLevelRemoteUrlReplacements != null ? userLevelRemoteUrlReplacements : "";
    }

    public static boolean isGlobbingEnabled(@NotNull Project project) {
        Boolean projectLevelIsGlobbingEnabled = getProjectLevelConfig(project).isGlobbingEnabled();
        return projectLevelIsGlobbingEnabled != null ? projectLevelIsGlobbingEnabled : getApplicationLevelConfig().isGlobbingEnabled();
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

    @NotNull
    @Contract(pure = true)
    public static String getPluginVersion() {
        IdeaPluginDescriptor plugin = PluginManagerCore.getPlugin(PluginId.getId("com.sourcegraph.jetbrains"));
        return plugin != null ? plugin.getVersion() : "unknown";
    }

    @Nullable
    public static String getAnonymousUserId() {
        return SourcegraphApplicationService.getInstance().getAnonymousUserId();
    }

    public static void setAnonymousUserId(@Nullable String anonymousUserId) {
        SourcegraphApplicationService.getInstance().anonymousUserId = anonymousUserId;
    }

    public static boolean isInstallEventLogged() {
        return SourcegraphApplicationService.getInstance().isInstallEventLogged();
    }

    public static void setInstallEventLogged(boolean value) {
        SourcegraphApplicationService.getInstance().isInstallEventLogged = value;
    }

    @NotNull
    private static SourcegraphApplicationService getApplicationLevelConfig() {
        return Objects.requireNonNull(SourcegraphApplicationService.getInstance());
    }

    @NotNull
    private static SourcegraphProjectService getProjectLevelConfig(@NotNull Project project) {
        return Objects.requireNonNull(SourcegraphProjectService.getInstance(project));
    }
}
