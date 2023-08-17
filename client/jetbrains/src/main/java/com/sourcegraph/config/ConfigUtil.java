package com.sourcegraph.config;

import com.google.gson.JsonObject;
import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginManagerCore;
import com.intellij.openapi.extensions.PluginId;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.agent.ExtensionConfiguration;
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
  public static ExtensionConfiguration getAgentConfiguration(@NotNull Project project) {
    return new ExtensionConfiguration()
        .setServerEndpoint(getSourcegraphUrl(project))
        .setAccessToken(getProjectAccessToken(project))
        .setCustomHeaders(getCustomRequestHeadersAsMap(project))
        .setAutocompleteAdvancedProvider(
            UserLevelConfig.getAutoCompleteProviderType().vscodeSettingString())
        .setAutocompleteAdvancedServerEndpoint(UserLevelConfig.getAutoCompleteServerEndpoint())
        .setAutocompleteAdvancedAccessToken(UserLevelConfig.getAutoCompleteAccessToken())
        .setAutocompleteAdvancedEmbeddings(UserLevelConfig.getAutocompleteAdvancedEmbeddings())
        .setDebug(isCodyDebugEnabled())
        .setVerboseDebug(isCodyVerboseDebugEnabled());
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
    if (projectLevelUrl != null && !projectLevelUrl.isEmpty()) {
      return addSlashIfNeeded(projectLevelUrl);
    }

    // Application level
    String applicationLevelUrl = getApplicationLevelConfig().getSourcegraphUrl();
    if (applicationLevelUrl != null && !applicationLevelUrl.isEmpty()) {
      return addSlashIfNeeded(applicationLevelUrl);
    }

    // User level or default
    String userLevelUrl = UserLevelConfig.getSourcegraphUrl();
    return !userLevelUrl.isEmpty() ? addSlashIfNeeded(userLevelUrl) : "";
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
    if (projectLevelCustomRequestHeaders != null && !projectLevelCustomRequestHeaders.isEmpty()) {
      return projectLevelCustomRequestHeaders;
    }

    // Application level
    String applicationLevelCustomRequestHeaders =
        getApplicationLevelConfig().getCustomRequestHeaders();
    if (applicationLevelCustomRequestHeaders != null
        && !applicationLevelCustomRequestHeaders.isEmpty()) {
      return applicationLevelCustomRequestHeaders;
    }

    // Default
    return "";
  }

  @NotNull
  public static String getDefaultBranchName(@NotNull Project project) {
    // Project level
    String projectLevelDefaultBranchName = getProjectLevelConfig(project).getDefaultBranchName();
    if (projectLevelDefaultBranchName != null && !projectLevelDefaultBranchName.isEmpty()) {
      return projectLevelDefaultBranchName;
    }

    // Application level
    String applicationLevelDefaultBranchName = getApplicationLevelConfig().getDefaultBranchName();
    if (applicationLevelDefaultBranchName != null && !applicationLevelDefaultBranchName.isEmpty()) {
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
    if (projectLevelReplacements != null && !projectLevelReplacements.isEmpty()) {
      return projectLevelReplacements;
    }

    // Application level
    String applicationLevelReplacements = getApplicationLevelConfig().getRemoteUrlReplacements();
    if (applicationLevelReplacements != null && !applicationLevelReplacements.isEmpty()) {
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
    return getApplicationLevelConfig().isCodyEnabled;
  }

  public static boolean isCodyDebugEnabled() {
    return getApplicationLevelConfig().isCodyDebugEnabled();
  }

  public static boolean isCodyVerboseDebugEnabled() {
    return getApplicationLevelConfig().isCodyVerboseDebugEnabled();
  }

  public static boolean isCodyAutoCompleteEnabled() {
    return getApplicationLevelConfig().isCodyAutoCompleteEnabled();
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

  // Null means user denied access to token storage. Empty string means no token found.
  @Nullable
  public static String getProjectAccessToken(@NotNull Project project) {
    SettingsComponent.InstanceType instanceType = ConfigUtil.getInstanceType(project);
    if (instanceType == SettingsComponent.InstanceType.ENTERPRISE) {
      return getEnterpriseAccessToken(project);
    } else if (instanceType == SettingsComponent.InstanceType.LOCAL_APP) {
      return LocalAppManager.getLocalAppAccessToken().orElse("");
    } else {
      return getDotComAccessToken(project);
    }
  }

  // Null means user denied access to token storage. Empty string means no token found.
  @Nullable
  public static String getEnterpriseAccessToken(Project project) {
    // Project level overrides secure storage
    String unsafeProjectLevelAccessToken =
        getProjectLevelConfig(project).getEnterpriseAccessToken();
    if (unsafeProjectLevelAccessToken != null) {
      return unsafeProjectLevelAccessToken;
    }

    // Get token from secure storage
    Optional<String> securelyStoredAccessToken = AccessTokenStorage.getEnterpriseAccessToken();
    if (securelyStoredAccessToken.isEmpty()) {
      return null; // Uer denied access to token storage
    }
    if (!securelyStoredAccessToken.get().isEmpty()) {
      return securelyStoredAccessToken.get();
    }

    // No secure token found, so use app-level token and migrate it to secure storage.
    String unsafeApplicationLevelAccessToken = getApplicationLevelConfig().enterpriseAccessToken;
    if (unsafeApplicationLevelAccessToken != null) {
      AccessTokenStorage.setApplicationEnterpriseAccessToken(unsafeApplicationLevelAccessToken);
      getApplicationLevelConfig().enterpriseAccessToken = null;
    }
    return unsafeApplicationLevelAccessToken != null ? unsafeApplicationLevelAccessToken : "";
  }

  // Null means user denied access to token storage. Empty string means no token found.
  @Nullable
  public static String getDotComAccessToken(@NotNull Project project) {
    // Project level overrides secure storage
    String projectLevelAccessToken = getProjectLevelConfig(project).getDotComAccessToken();
    if (projectLevelAccessToken != null) {
      return projectLevelAccessToken;
    }

    // Get token from secure storage
    Optional<String> securelyStoredAccessToken = AccessTokenStorage.getDotComAccessToken();
    if (securelyStoredAccessToken.isEmpty()) {
      return null; // Uer denied access to token storage
    }
    if (!securelyStoredAccessToken.get().isEmpty()) {
      return securelyStoredAccessToken.get();
    }

    // No secure token found, so use app-level token and migrate it to secure storage.
    String unsafeApplicationLevelAccessToken = getApplicationLevelConfig().dotComAccessToken;
    if (unsafeApplicationLevelAccessToken != null) {
      AccessTokenStorage.setApplicationDotComAccessToken(unsafeApplicationLevelAccessToken);
      getApplicationLevelConfig().dotComAccessToken = null;
    }
    return unsafeApplicationLevelAccessToken != null ? unsafeApplicationLevelAccessToken : "";
  }
}
