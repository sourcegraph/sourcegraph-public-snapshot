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
    public static final String DOTCOM_URL = "https://sourcegraph.com/";

    @NotNull
    public static JsonObject getConfigAsJson(@NotNull Project project) {
        JsonObject configAsJson = new JsonObject();
        configAsJson.addProperty("instanceURL", ConfigUtil.getSourcegraphUrl(project));
        configAsJson.addProperty("accessToken", ConfigUtil.getInstanceType(project) == SettingsComponent.InstanceType.ENTERPRISE ? ConfigUtil.getAccessToken(project) : null);
        configAsJson.addProperty("customRequestHeadersAsString", ConfigUtil.getCustomRequestHeaders(project));
        configAsJson.addProperty("isGlobbingEnabled", ConfigUtil.isGlobbingEnabled(project));
        configAsJson.addProperty("pluginVersion", ConfigUtil.getPluginVersion());
        configAsJson.addProperty("anonymousUserId", ConfigUtil.getAnonymousUserId());
        return configAsJson;
    }

    @NotNull
    public static SettingsComponent.InstanceType getInstanceType(Project project) {
        // Project level
        String projectLevelSetting = getProjectLevelConfig(project).getInstanceType();
        if (projectLevelSetting != null && !projectLevelSetting.isEmpty()) {
            return projectLevelSetting.equals(SettingsComponent.InstanceType.ENTERPRISE.name())
                ? SettingsComponent.InstanceType.ENTERPRISE : SettingsComponent.InstanceType.DOTCOM;
        }


        // Application level
        String applicationLevelSetting = getApplicationLevelConfig().getInstanceType();
        if (applicationLevelSetting != null && !applicationLevelSetting.isEmpty()) {
            return applicationLevelSetting.equals(SettingsComponent.InstanceType.ENTERPRISE.name())
                ? SettingsComponent.InstanceType.ENTERPRISE : SettingsComponent.InstanceType.DOTCOM;
        }

        // User level or default
        String enterpriseUrl = getEnterpriseUrl(project);
        return (enterpriseUrl.equals("") || enterpriseUrl.startsWith(DOTCOM_URL))
            ? SettingsComponent.InstanceType.DOTCOM : SettingsComponent.InstanceType.ENTERPRISE;
    }

    @NotNull
    public static String getSourcegraphUrl(@NotNull Project project) {
        if (getInstanceType(project) == SettingsComponent.InstanceType.DOTCOM) {
            return DOTCOM_URL;
        } else {
            String enterpriseUrl = getEnterpriseUrl(project);
            return !enterpriseUrl.isEmpty() ? enterpriseUrl : DOTCOM_URL;
        }
    }

    @NotNull
    public static String getEnterpriseUrl(@NotNull Project project) {
        // Project level
        String projectLevelUrl = getProjectLevelConfig(project).getSourcegraphUrl();
        if (projectLevelUrl != null && projectLevelUrl.length() > 0) {
            return addSlashIfNeeded(projectLevelUrl);
        }

        // Application level
        String applicationLevelUrl = getApplicationLevelConfig().getSourcegraphUrl();
        if (applicationLevelUrl != null && applicationLevelUrl.length() > 0) {
            return addSlashIfNeeded(applicationLevelUrl);
        }

        // User level or default
        String userLevelUrl = UserLevelConfig.getSourcegraphUrl();
        return !userLevelUrl.equals("") ? addSlashIfNeeded(userLevelUrl) : "";
    }

    @Nullable
    public static String getAccessToken(Project project) {
        // Project level → application level
        String projectLevelAccessToken = getProjectLevelConfig(project).getAccessToken();
        return projectLevelAccessToken != null ? projectLevelAccessToken : getApplicationLevelConfig().getAccessToken();
    }

    @NotNull
    public static String getCustomRequestHeaders(@NotNull Project project) {
        // Project level
        String projectLevelCustomRequestHeaders = getProjectLevelConfig(project).getCustomRequestHeaders();
        if (projectLevelCustomRequestHeaders != null && projectLevelCustomRequestHeaders.length() > 0) {
            return projectLevelCustomRequestHeaders;
        }

        // Application level
        String applicationLevelCustomRequestHeaders = getApplicationLevelConfig().getCustomRequestHeaders();
        if (applicationLevelCustomRequestHeaders != null && applicationLevelCustomRequestHeaders.length() > 0) {
            return applicationLevelCustomRequestHeaders;
        }

        // Default
        return "";
    }

    @NotNull
    public static String getDefaultBranchName(@NotNull Project project) {
        // Project level
        String projectLevelDefaultBranchName = getProjectLevelConfig(project).getDefaultBranchName();
        if (projectLevelDefaultBranchName != null && projectLevelDefaultBranchName.length() > 0) {
            return projectLevelDefaultBranchName;
        }

        // Application level
        String applicationLevelDefaultBranchName = getApplicationLevelConfig().getDefaultBranchName();
        if (applicationLevelDefaultBranchName != null && applicationLevelDefaultBranchName.length() > 0) {
            return applicationLevelDefaultBranchName;
        }

        // User level or default
        String userLevelDefaultBranchName = UserLevelConfig.getDefaultBranchName();
        return userLevelDefaultBranchName != null ? userLevelDefaultBranchName : "main";
    }

    @NotNull
    public static String getRemoteUrlReplacements(@NotNull Project project) {
        // Project level
        String projectLevelReplacements = getProjectLevelConfig(project).getRemoteUrlReplacements();
        if (projectLevelReplacements != null && projectLevelReplacements.length() > 0) {
            return projectLevelReplacements;
        }

        // Application level
        String applicationLevelReplacements = getApplicationLevelConfig().getRemoteUrlReplacements();
        if (applicationLevelReplacements != null && applicationLevelReplacements.length() > 0) {
            return applicationLevelReplacements;
        }

        // User level or default
        String userLevelRemoteUrlReplacements = UserLevelConfig.getRemoteUrlReplacements();
        return userLevelRemoteUrlReplacements != null ? userLevelRemoteUrlReplacements : "";
    }

    public static boolean isGlobbingEnabled(@NotNull Project project) {
        // Project level → application level
        Boolean projectLevelIsGlobbingEnabled = getProjectLevelConfig(project).isGlobbingEnabled();
        return projectLevelIsGlobbingEnabled != null ? projectLevelIsGlobbingEnabled : getApplicationLevelConfig().isGlobbingEnabled();
    }

    @Nullable
    public static Search getLastSearch(@NotNull Project project) {
        // Project level
        return getProjectLevelConfig(project).getLastSearch();
    }

    public static void setLastSearch(@NotNull Project project, @NotNull Search lastSearch) {
        // Project level
        SourcegraphProjectService settings = getProjectLevelConfig(project);
        settings.lastSearchQuery = lastSearch.getQuery() != null ? lastSearch.getQuery() : "";
        settings.lastSearchCaseSensitive = lastSearch.isCaseSensitive();
        settings.lastSearchPatternType = lastSearch.getPatternType() != null ? lastSearch.getPatternType() : "literal";
        settings.lastSearchContextSpec = lastSearch.getSelectedSearchContextSpec() != null ? lastSearch.getSelectedSearchContextSpec() : "global";
    }

    @NotNull
    @Contract(pure = true)
    public static String getPluginVersion() {
        // Internal version
        IdeaPluginDescriptor plugin = PluginManagerCore.getPlugin(PluginId.getId("com.sourcegraph.jetbrains"));
        return plugin != null ? plugin.getVersion() : "unknown";
    }

    @Nullable
    public static String getAnonymousUserId() {
        return getApplicationLevelConfig().getAnonymousUserId();
    }

    public static void setAnonymousUserId(@Nullable String anonymousUserId) {
        getApplicationLevelConfig().anonymousUserId = anonymousUserId;
    }

    public static boolean isInstallEventLogged() {
        return getApplicationLevelConfig().isInstallEventLogged();
    }

    public static void setInstallEventLogged(boolean value) {
        getApplicationLevelConfig().isInstallEventLogged = value;
    }

    public static boolean isUrlNotificationDismissed() {
        return getApplicationLevelConfig().isUrlNotificationDismissed();
    }

    public static void setUrlNotificationDismissed(boolean value) {
        getApplicationLevelConfig().isUrlNotificationDismissed = value;
    }

    @NotNull
    private static String addSlashIfNeeded(@NotNull String url) {
        return url.endsWith("/") ? url : url + "/";
    }

    public static boolean didAuthenticationFailLastTime() {
        Boolean failedLastTime = getApplicationLevelConfig().getAuthenticationFailedLastTime();
        return failedLastTime != null ? failedLastTime : true;
    }

    public static void setAuthenticationFailedLastTime(boolean value) {
        SourcegraphApplicationService.getInstance().authenticationFailedLastTime = value;
    }

    public static String getLastUpdateNotificationPluginVersion() {
        return SourcegraphApplicationService.getInstance().getLastUpdateNotificationPluginVersion();
    }

    public static void setLastUpdateNotificationPluginVersionToCurrent() {
        SourcegraphApplicationService.getInstance().lastUpdateNotificationPluginVersion = getPluginVersion();
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
