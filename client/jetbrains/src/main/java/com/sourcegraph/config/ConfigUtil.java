package com.sourcegraph.config;

import com.google.gson.JsonObject;
import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginManagerCore;
import com.intellij.openapi.extensions.PluginId;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.agent.ConnectionConfiguration;
import com.sourcegraph.cody.localapp.LocalAppManager;
import com.sourcegraph.find.Search;
import java.util.HashMap;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.Contract;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ConfigUtil {
  public static final String DOTCOM_URL = "https://sourcegraph.com/";

  @NotNull
  public static ConnectionConfiguration getAgentConfiguration(@NotNull Project project) {
    return new ConnectionConfiguration()
        .setServerEndpoint(getSourcegraphUrl(project))
        .setAccessToken(getProjectAccessToken(project))
        .setCustomHeaders(getCustomRequestHeadersAsMap(project));
  }

  @NotNull
  public static JsonObject getConfigAsJson(@NotNull Project project) {
    JsonObject configAsJson = new JsonObject();
    configAsJson.addProperty("instanceURL", ConfigUtil.getSourcegraphUrl(project));
    configAsJson.addProperty("accessToken", ConfigUtil.getProjectAccessToken(project));
    configAsJson.addProperty(
        "customRequestHeadersAsString", ConfigUtil.getCustomRequestHeaders(project));
    configAsJson.addProperty("pluginVersion", ConfigUtil.getPluginVersion());
    configAsJson.addProperty("anonymousUserId", ConfigUtil.getAnonymousUserId());
    return configAsJson;
  }

  @NotNull
  public static SettingsComponent.InstanceType getInstanceType(@NotNull Project project) {
    return Optional.ofNullable(getProjectLevelConfig(project).getInstanceType()) // Project level
        .flatMap(SettingsComponent.InstanceType::optionalValueOf)
        .or( // Application level
            () ->
                Optional.ofNullable(getApplicationLevelConfig().getInstanceType())
                    .flatMap(SettingsComponent.InstanceType::optionalValueOf))
        .or( // User level
            () ->
                Optional.of(getEnterpriseUrl(project))
                    .filter(StringUtils::isNotEmpty)
                    .flatMap(
                        url -> {
                          if (url.startsWith(DOTCOM_URL)) {
                            return Optional.of(SettingsComponent.InstanceType.DOTCOM);
                          } else if (url.startsWith(LocalAppManager.getLocalAppUrl())) {
                            return Optional.of(SettingsComponent.InstanceType.LOCAL_APP);
                          } else {
                            return Optional.empty();
                          }
                        }))
        .orElse(SettingsComponent.getDefaultInstanceType()); // or default
  }

  @NotNull
  public static String getSourcegraphUrl(@NotNull Project project) {
    SettingsComponent.InstanceType instanceType = getInstanceType(project);
    if (instanceType == SettingsComponent.InstanceType.DOTCOM) {
      return DOTCOM_URL;
    } else if (instanceType == SettingsComponent.InstanceType.LOCAL_APP) {
      return LocalAppManager.getLocalAppUrl();
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

  public static Map<String, String> getCustomRequestHeadersAsMap(@NotNull Project project) {
    Map<String, String> result = new HashMap<>();
    String[] pairs = getCustomRequestHeaders(project).split(",");
    for (int i = 0; i + 1 < pairs.length; i = i + 2) {
      result.put(pairs[i], pairs[i + 1]);
    }
    return result;
  }

  @NotNull
  public static String getCustomRequestHeaders(@NotNull Project project) {
    // Project level
    String projectLevelCustomRequestHeaders =
        getProjectLevelConfig(project).getCustomRequestHeaders();
    if (projectLevelCustomRequestHeaders != null && projectLevelCustomRequestHeaders.length() > 0) {
      return projectLevelCustomRequestHeaders;
    }

    // Application level
    String applicationLevelCustomRequestHeaders =
        getApplicationLevelConfig().getCustomRequestHeaders();
    if (applicationLevelCustomRequestHeaders != null
        && applicationLevelCustomRequestHeaders.length() > 0) {
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
    if (applicationLevelDefaultBranchName != null
        && applicationLevelDefaultBranchName.length() > 0) {
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

  @Nullable
  public static Search getLastSearch(@NotNull Project project) {
    // Project level
    return getProjectLevelConfig(project).getLastSearch();
  }

  public static void setLastSearch(@NotNull Project project, @NotNull Search lastSearch) {
    // Project level
    CodyProjectService settings = getProjectLevelConfig(project);
    settings.lastSearchQuery = lastSearch.getQuery() != null ? lastSearch.getQuery() : "";
    settings.lastSearchCaseSensitive = lastSearch.isCaseSensitive();
    settings.lastSearchPatternType =
        lastSearch.getPatternType() != null ? lastSearch.getPatternType() : "literal";
    settings.lastSearchContextSpec =
        lastSearch.getSelectedSearchContextSpec() != null
            ? lastSearch.getSelectedSearchContextSpec()
            : "global";
  }

  @NotNull
  @Contract(pure = true)
  public static String getPluginVersion() {
    // Internal version
    IdeaPluginDescriptor plugin =
        PluginManagerCore.getPlugin(PluginId.getId("com.sourcegraph.jetbrains"));
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

  public static boolean isCodyEnabled() {
    return getApplicationLevelConfig().isCodyEnabled();
  }

  public static boolean isCodyAutoCompleteEnabled() {
    return getApplicationLevelConfig().isCodyEnabled()
        && getApplicationLevelConfig().isCodyAutoCompleteEnabled();
  }

  public static boolean isAccessTokenNotificationDismissed() {
    return getApplicationLevelConfig().isAccessTokenNotificationDismissed();
  }

  public static void setUrlNotificationDismissed(boolean value) {
    getApplicationLevelConfig().isUrlNotificationDismissed = value;
  }

  public static void setAccessTokenNotificationDismissed(boolean value) {
    getApplicationLevelConfig().isAccessTokenNotificationDismissed = value;
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
    CodyApplicationService.getInstance().authenticationFailedLastTime = value;
  }

  public static String getLastUpdateNotificationPluginVersion() {
    return CodyApplicationService.getInstance().getLastUpdateNotificationPluginVersion();
  }

  public static void setLastUpdateNotificationPluginVersionToCurrent() {
    CodyApplicationService.getInstance().lastUpdateNotificationPluginVersion = getPluginVersion();
  }

  @NotNull
  private static CodyApplicationService getApplicationLevelConfig() {
    return Objects.requireNonNull(CodyApplicationService.getInstance());
  }

  @NotNull
  private static CodyProjectService getProjectLevelConfig(@NotNull Project project) {
    return Objects.requireNonNull(CodyProjectService.getInstance(project));
  }

  @Nullable
  public static String getWorkspaceRoot(@NotNull Project project) {
    if (project.getBasePath() != null) {
      return project.getBasePath();
    }
    // The base path should only be null for the default project. The agent server assumes that the
    // workspace root is not null, so we have to provide some default value. Feel free to change to
    // something else than the home directory if this is causing problems.
    return System.getProperty("user.home");
  }

  @Nullable
  public static String getProjectAccessToken(@NotNull Project project) {
    SettingsComponent.InstanceType instanceType = ConfigUtil.getInstanceType(project);
    if (instanceType == SettingsComponent.InstanceType.ENTERPRISE) {
      return getEnterpriseAccessToken(project);
    } else if (instanceType == SettingsComponent.InstanceType.LOCAL_APP) {
      return LocalAppManager.getLocalAppAccessToken().orElse(null);
    } else {
      return getDotComAccessToken(project);
    }
  }

  @Nullable
  public static String getEnterpriseAccessToken(Project project) {
    // project level → application level secure storage -> application level
    return Optional.ofNullable(getProjectLevelConfig(project).getEnterpriseAccessToken())
        .or(AccessTokenStorage::getEnterpriseAccessToken)
        .orElseGet(
            () -> {
              // Save the application level access token to the secure storage
              String unsafeApplicationLevelAccessToken =
                  getApplicationLevelConfig().getEnterpriseAccessToken();
              if (unsafeApplicationLevelAccessToken != null) {
                AccessTokenStorage.setApplicationEnterpriseAccessToken(
                    unsafeApplicationLevelAccessToken);
              }
              return unsafeApplicationLevelAccessToken;
            });
  }

  @Nullable
  public static String getDotComAccessToken(@NotNull Project project) {
    // project level → application level secure storage -> application level
    return Optional.ofNullable(getProjectLevelConfig(project).getDotComAccessToken())
        .or(AccessTokenStorage::getDotComAccessToken)
        .orElseGet(
            () -> {
              // Save the application level access token to the secure storage
              String unsafeApplicationLevelAccessToken =
                  getApplicationLevelConfig().getDotComAccessToken();
              if (unsafeApplicationLevelAccessToken != null) {
                AccessTokenStorage.setApplicationDotComAccessToken(
                    unsafeApplicationLevelAccessToken);
              }
              return unsafeApplicationLevelAccessToken;
            });
  }
}
