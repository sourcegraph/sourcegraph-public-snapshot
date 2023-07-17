package com.sourcegraph.cody.api;

import com.google.gson.JsonObject;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.api.GraphQlClient;
import com.sourcegraph.api.GraphQlResponse;
import com.sourcegraph.config.ConfigUtil;
import java.util.Optional;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;
import org.jetbrains.annotations.NotNull;

public class CodyLLMConfiguration {
  public static Logger logger = Logger.getInstance(CodyLLMConfiguration.class);

  // TODO: consolidate and reuse executors in the plugin
  private final ExecutorService executor = Executors.newSingleThreadExecutor();
  public static final int DEFAULT_CHAT_MODEL_MAX_TOKENS = 7000;
  private final @NotNull Project project;

  private final AtomicInteger chatModelMaxTokensCache = new AtomicInteger();

  public CodyLLMConfiguration(@NotNull Project project) {
    this.project = project;
  }

  public static CodyLLMConfiguration getInstance(@NotNull Project project) {
    return project.getService(CodyLLMConfiguration.class);
  }

  public int getChatModelMaxTokensForProject() {
    if (chatModelMaxTokensCache.get() > 0) {
      return chatModelMaxTokensCache.get();
    } else {
      // schedules a cache refresh and uses the default value if the cache was empty
      refreshCache();
      return DEFAULT_CHAT_MODEL_MAX_TOKENS;
    }
  }

  public void refreshCache() {
    this.executor.submit(() -> chatModelMaxTokensCache.set(fetchChatModelMaxTokens().orElse(0)));
  }

  private Optional<Integer> fetchChatModelMaxTokens() {
    String graphQlQuery =
        "query fetchChatModelMaxTokens {\n"
            + "  site {\n"
            + "    codyLLMConfiguration {\n"
            + "      chatModelMaxTokens\n"
            // + "      completionModelMaxTokens\n" // TODO apply this for autocomplete later
            + "    }\n"
            + "  }\n"
            + "}";
    String instanceUrl = ConfigUtil.getSourcegraphUrl(this.project);
    String accessToken = ConfigUtil.getProjectAccessToken(this.project);
    String customRequestHeaders = ConfigUtil.getCustomRequestHeaders(this.project);
    try {
      GraphQlResponse response =
          GraphQlClient.callGraphQLService(
              instanceUrl, accessToken, customRequestHeaders, graphQlQuery, new JsonObject());
      JsonObject body = response.getBodyAsJson();
      if (body.has("errors")) {
        logger.warn("Fetching chat model max tokens failed with errors: " + body.get("errors"));
        return Optional.empty();
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
                  logger.warn("Failed to fetch a valid value for chat model max tokens to");
                  return Optional.empty();
                }
              });
    } catch (Exception e) {
      logger.warn(e);
      logger.warn("Could not fetch chat model max tokens from Sourcegraph instance");
      return Optional.empty();
    }
  }
}
