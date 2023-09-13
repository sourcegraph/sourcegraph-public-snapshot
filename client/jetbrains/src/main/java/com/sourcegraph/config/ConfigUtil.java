package com.sourcegraph.config;

import com.google.gson.JsonObject;
import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginManagerCore;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.extensions.PluginId;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.project.ProjectManager;
import com.sourcegraph.cody.agent.ExtensionConfiguration;
import com.sourcegraph.cody.config.CodyAccount;
import com.sourcegraph.cody.config.CodyApplicationSettings;
import com.sourcegraph.cody.config.CodyAuthenticationManager;
import com.sourcegraph.cody.config.ServerAuth;
import com.sourcegraph.cody.config.ServerAuthLoader;
import com.sourcegraph.cody.config.SourcegraphServerPath;
import java.util.*;
import java.util.stream.Collectors;
import org.jetbrains.annotations.Contract;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ConfigUtil {
  public static final String DOTCOM_URL = "https://sourcegraph.com/";
  public static final String SERVICE_DISPLAY_NAME = "Sourcegraph Cody + Code Search";

  @NotNull
  public static ExtensionConfiguration getAgentConfiguration(@NotNull Project project) {
    ServerAuth serverAuth = ServerAuthLoader.loadServerAuth(project);
    ExtensionConfiguration config =
        new ExtensionConfiguration()
            .setServerEndpoint(serverAuth.getInstanceUrl())
            .setAccessToken(serverAuth.getAccessToken())
            .setCustomHeaders(getCustomRequestHeadersAsMap(serverAuth.getCustomRequestHeaders()))
            .setProxy(UserLevelConfig.getProxy())
            .setAutocompleteAdvancedServerEndpoint(UserLevelConfig.getAutocompleteServerEndpoint())
            .setAutocompleteAdvancedAccessToken(UserLevelConfig.getAutocompleteAccessToken())
            .setAutocompleteAdvancedEmbeddings(UserLevelConfig.getAutocompleteAdvancedEmbeddings())
            .setDebug(isCodyDebugEnabled())
            .setVerboseDebug(isCodyVerboseDebugEnabled());

    if (UserLevelConfig.getAutocompleteProviderType() != null) {
      config.setAutocompleteAdvancedProvider(
          UserLevelConfig.getAutocompleteProviderType().vscodeSettingString());
    }
    return config;
  }

  @NotNull
  public static JsonObject getConfigAsJson(@NotNull Project project) {
    JsonObject configAsJson = new JsonObject();
    ServerAuth serverAuth = ServerAuthLoader.loadServerAuth(project);
    CodyApplicationSettings codyApplicationSettings = CodyApplicationSettings.getInstance();
    configAsJson.addProperty("instanceURL", serverAuth.getInstanceUrl());
    configAsJson.addProperty("accessToken", serverAuth.getAccessToken());
    configAsJson.addProperty("customRequestHeadersAsString", serverAuth.getCustomRequestHeaders());
    configAsJson.addProperty("pluginVersion", ConfigUtil.getPluginVersion());
    configAsJson.addProperty("anonymousUserId", codyApplicationSettings.getAnonymousUserId());
    return configAsJson;
  }

  @NotNull
  public static SourcegraphServerPath getServerPath(@NotNull Project project) {
    CodyAccount defaultAccount = CodyAuthenticationManager.getInstance().getDefaultAccount(project);

    return defaultAccount != null
        ? defaultAccount.getServer()
        : SourcegraphServerPath.from(DOTCOM_URL, "");
  }

  public static Map<String, String> getCustomRequestHeadersAsMap(
      @NotNull String customRequestHeaders) {
    Map<String, String> result = new HashMap<>();
    String[] pairs = customRequestHeaders.split(",");
    for (int i = 0; i + 1 < pairs.length; i = i + 2) {
      result.put(pairs[i], pairs[i + 1]);
    }
    return result;
  }

  @NotNull
  @Contract(pure = true)
  public static String getPluginVersion() {
    // Internal version
    IdeaPluginDescriptor plugin =
        PluginManagerCore.getPlugin(PluginId.getId("com.sourcegraph.jetbrains"));
    return plugin != null ? plugin.getVersion() : "unknown";
  }

  public static boolean isDefaultDotcomAccountNotificationDismissed() {
    return CodyApplicationSettings.getInstance().isDefaultDotcomAccountNotificationDismissed();
  }

  public static boolean isCodyEnabled() {
    return CodyApplicationSettings.getInstance().isCodyEnabled();
  }

  public static boolean isCodyDebugEnabled() {
    return CodyApplicationSettings.getInstance().isCodyDebugEnabled();
  }

  public static boolean isCodyVerboseDebugEnabled() {
    return CodyApplicationSettings.getInstance().isCodyVerboseDebugEnabled();
  }

  public static boolean isCodyAutocompleteEnabled() {
    return CodyApplicationSettings.getInstance().isCodyAutocompleteEnabled();
  }

  public static boolean isCustomAutocompleteColorEnabled() {
    return CodyApplicationSettings.getInstance().isCustomAutocompleteColorEnabled();
  }

  public static Integer getCustomAutocompleteColor() {
    return CodyApplicationSettings.getInstance().getCustomAutocompleteColor();
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

  public static List<Editor> getAllEditors() {
    Project[] openProjects = ProjectManager.getInstance().getOpenProjects();
    return Arrays.stream(openProjects)
        .flatMap(project -> Arrays.stream(FileEditorManager.getInstance(project).getAllEditors()))
        .filter(fileEditor -> fileEditor instanceof com.intellij.openapi.fileEditor.TextEditor)
        .map(fileEditor -> ((com.intellij.openapi.fileEditor.TextEditor) fileEditor).getEditor())
        .collect(Collectors.toList());
  }

  public static List<String> getBlacklistedAutocompleteLanguageIds() {
    return CodyApplicationSettings.getInstance().getBlacklistedLanguageIds();
  }
}
