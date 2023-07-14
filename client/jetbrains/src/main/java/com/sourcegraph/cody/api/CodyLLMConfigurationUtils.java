package com.sourcegraph.cody.api;

import com.google.gson.JsonObject;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.api.GraphQlClient;
import com.sourcegraph.api.GraphQlResponse;
import com.sourcegraph.config.ConfigUtil;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import org.jetbrains.annotations.NotNull;

public class CodyLLMConfigurationUtils {
  public static Logger logger = Logger.getInstance(CodyLLMConfigurationUtils.class);
  public static final int DEFAULT_CHAT_MODEL_MAX_TOKENS = 7000;
  private static final ConcurrentHashMap<String, Integer> projectNameToChatModelMaxTokensCache =
      new ConcurrentHashMap<>();

  public static int getChatModelMaxTokensForProject(@NotNull Project project) {
    String projectName = project.getName();
    return Optional.ofNullable(projectNameToChatModelMaxTokensCache.get(projectName))
        .or(
            () -> {
              int newValue = fetchChatModelMaxTokens(project);
              projectNameToChatModelMaxTokensCache.put(projectName, newValue);
              return Optional.of(newValue);
            })
        .orElse(DEFAULT_CHAT_MODEL_MAX_TOKENS);
  }

  public static void refreshCacheForProject(@NotNull Project project) {
    projectNameToChatModelMaxTokensCache.put(project.getName(), fetchChatModelMaxTokens(project));
  }

  private static int fetchChatModelMaxTokens(@NotNull Project project) {
    String graphQlQuery =
        "query fetchChatModelMaxTokens {\n"
            + "  site {\n"
            + "    codyLLMConfiguration {\n"
            + "      chatModelMaxTokens\n"
            // + "      completionModelMaxTokens\n" // TODO apply this for autocomplete later
            + "    }\n"
            + "  }\n"
            + "}";
    String instanceUrl = ConfigUtil.getSourcegraphUrl(project);
    String accessToken = ConfigUtil.getProjectAccessToken(project);
    String customRequestHeaders = ConfigUtil.getCustomRequestHeaders(project);
    try {
      GraphQlResponse response =
          GraphQlClient.callGraphQLService(
              instanceUrl, accessToken, customRequestHeaders, graphQlQuery, new JsonObject());
      JsonObject body = response.getBodyAsJson();
      if (body.has("errors")) {
        logger.warn("Fetching chat model max tokens failed with errors: " + body.get("errors"));
        logger.warn("Defaulting chat model max tokens to: " + DEFAULT_CHAT_MODEL_MAX_TOKENS);
        return DEFAULT_CHAT_MODEL_MAX_TOKENS;
      }
      return Optional.ofNullable(body.getAsJsonObject("data"))
          .flatMap(data -> Optional.ofNullable(data.getAsJsonObject("site")))
          .flatMap(site -> Optional.ofNullable(site.getAsJsonObject("codyLLMConfiguration")))
          .flatMap(
              codyLLMConfiguration ->
                  Optional.ofNullable(
                      codyLLMConfiguration.getAsJsonPrimitive(("chatModelMaxTokens"))))
          .flatMap(
              r -> {
                try {
                  return Optional.of(r.getAsInt());
                } catch (NumberFormatException e) {
                  logger.warn(e);
                  logger.warn(
                      "Failed to fetch a valid value, defaulting chat model max tokens to: "
                          + DEFAULT_CHAT_MODEL_MAX_TOKENS);
                  return Optional.of(DEFAULT_CHAT_MODEL_MAX_TOKENS);
                }
              })
          .orElse(DEFAULT_CHAT_MODEL_MAX_TOKENS);
    } catch (Exception e) {

      logger.warn(e);
      logger.warn(
          "Could not fetch chat model max tokens from Sourcegraph instance, defaulting to: "
              + DEFAULT_CHAT_MODEL_MAX_TOKENS);
      return DEFAULT_CHAT_MODEL_MAX_TOKENS;
    }
  }
}
