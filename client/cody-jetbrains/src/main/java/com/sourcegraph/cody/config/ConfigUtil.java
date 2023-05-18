package com.sourcegraph.cody.config;

import com.intellij.openapi.project.Project;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.Objects;

public class ConfigUtil {
    public static final String DOTCOM_URL = "https://sourcegraph.com/";

    @NotNull
    public static SettingsComponent.InstanceType getInstanceType(@Nullable Project project) {
        // Project level
        if (project != null) {
            String projectLevelSetting = getProjectLevelConfig(project).getInstanceType();
            if (projectLevelSetting != null && !projectLevelSetting.isEmpty()) {
                return projectLevelSetting.equals(SettingsComponent.InstanceType.ENTERPRISE.name())
                    ? SettingsComponent.InstanceType.ENTERPRISE : SettingsComponent.InstanceType.DOTCOM;
            }
        }

        // Application level
        String applicationLevelSetting = getApplicationLevelConfig().getInstanceType();
        if (applicationLevelSetting != null && !applicationLevelSetting.isEmpty()) {
            return applicationLevelSetting.equals(SettingsComponent.InstanceType.ENTERPRISE.name())
                ? SettingsComponent.InstanceType.ENTERPRISE : SettingsComponent.InstanceType.DOTCOM;
        }

        // Use default
        String enterpriseUrl = getEnterpriseUrl(project);
        return (enterpriseUrl.equals("") || enterpriseUrl.startsWith(DOTCOM_URL))
            ? SettingsComponent.InstanceType.DOTCOM : SettingsComponent.InstanceType.ENTERPRISE;
    }

    @NotNull
    public static String getSourcegraphUrl(@Nullable Project project) {
        if (getInstanceType(project) == SettingsComponent.InstanceType.DOTCOM) {
            return DOTCOM_URL;
        } else {
            String enterpriseUrl = getEnterpriseUrl(project);
            return !enterpriseUrl.isEmpty() ? enterpriseUrl : DOTCOM_URL;
        }
    }

    @NotNull
    public static String getEnterpriseUrl(@Nullable Project project) {
        // Project level
        if (project != null) {
            String projectLevelUrl = getProjectLevelConfig(project).getEnterpriseUrl();
            if (projectLevelUrl != null && projectLevelUrl.length() > 0) {
                return addSlashIfNeeded(projectLevelUrl);
            }
        }

        // Application level
        String applicationLevelUrl = getApplicationLevelConfig().getEnterpriseUrl();
        if (applicationLevelUrl != null && applicationLevelUrl.length() > 0) {
            return addSlashIfNeeded(applicationLevelUrl);
        }

        // Use default
        return "";
    }

    @Nullable
    public static String getDotcomAccessToken(@Nullable Project project) {
        // Project level → application level
        String accessToken = project != null ? getProjectLevelConfig(project).getDotcomAccessToken() : null;
        return accessToken != null ? accessToken : getApplicationLevelConfig().getDotcomAccessToken();
    }

    @Nullable
    public static String getEnterpriseAccessToken(@Nullable Project project) {
        // Project level → application level
        String accessToken = project != null ? getProjectLevelConfig(project).getEnterpriseAccessToken() : null;
        return accessToken != null ? accessToken : getApplicationLevelConfig().getEnterpriseAccessToken();
    }

    @NotNull
    public static String getCustomRequestHeaders(@Nullable Project project) {
        // Project level
        if (project != null) {
            String projectLevelCustomRequestHeaders = getProjectLevelConfig(project).getCustomRequestHeaders();
            if (projectLevelCustomRequestHeaders != null && projectLevelCustomRequestHeaders.length() > 0) {
                return projectLevelCustomRequestHeaders;
            }
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
    public static String getCodebase(@Nullable Project project) {
        // Project level
        if (project != null) {
            String projectLevelCodebase = getProjectLevelConfig(project).getCodebase();
            if (projectLevelCodebase != null && projectLevelCodebase.length() > 0) {
                return projectLevelCodebase;
            }
        }

        // Application level
        String applicationLevelCodebase = getApplicationLevelConfig().getCodebase();
        if (applicationLevelCodebase != null && applicationLevelCodebase.length() > 0) {
            return applicationLevelCodebase;
        }

        // Use default
        return "";
    }

    @Nullable
    public static String getAnonymousUserId() {
        return getApplicationLevelConfig().getAnonymousUserId();
    }

    public static void setAnonymousUserId(@Nullable String anonymousUserId) {
        getApplicationLevelConfig().setAnonymousUserId(anonymousUserId);
    }

    public static boolean isInstallEventLogged() {
        return getApplicationLevelConfig().isInstallEventLogged();
    }

    public static void setInstallEventLogged(boolean value) {
        getApplicationLevelConfig().setInstallEventLogged(value);
    }

    public static boolean areChatPredictionsEnabled(@Nullable Project project) {
        // Project level → application level
        Boolean areChatPredictionsEnabled = project != null ? getProjectLevelConfig(project).areChatPredictionsEnabled() : null;
        return areChatPredictionsEnabled != null ? areChatPredictionsEnabled : Boolean.TRUE.equals(getApplicationLevelConfig().areChatPredictionsEnabled());
    }

    public static void setAreChatPredictionsEnabled(boolean value) {
        getApplicationLevelConfig().setChatPredictionsEnabled(value);
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
        CodyApplicationService.getInstance().setAuthenticationFailedLastTime(value);
    }

    @NotNull
    private static CodyApplicationService getApplicationLevelConfig() {
        return Objects.requireNonNull(CodyApplicationService.getInstance());
    }

    @NotNull
    private static CodyProjectService getProjectLevelConfig(@NotNull Project project) {
        return Objects.requireNonNull(CodyProjectService.getInstance(project));
    }
}
