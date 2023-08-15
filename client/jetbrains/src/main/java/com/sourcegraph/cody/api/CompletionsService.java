package com.sourcegraph.cody.api;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonObject;
import com.sourcegraph.api.GraphQlClient;
import java.io.IOException;
import org.jetbrains.annotations.NotNull;

public class CompletionsService {
  private final String instanceUrl;
  private final String accessToken;

  public CompletionsService(@NotNull String instanceUrl, @NotNull String accessToken) {
    this.instanceUrl = instanceUrl;
    this.accessToken = accessToken;
  }

  public enum Endpoint {
    Stream(".api/completions/stream"),
    Code(".api/completions/code");
    public final String urlPath;

    Endpoint(String urlPath) {
      this.urlPath = urlPath;
    }
  }

  /** Sends a completions request to the Sourcegraph instance, and returns the response. */
  public String getCompletion(@NotNull CompletionsInput input)
      throws IOException, InterruptedException {
    Gson gson =
        new GsonBuilder()
            .registerTypeAdapter(Speaker.class, new SpeakerLowercaseSerializer())
            .create();

    String query = "query completions($input: CompletionsInput!) { completions(input: $input) }";
    var variables = new JsonObject();
    variables.add("input", gson.toJsonTree(input));

    var response =
        GraphQlClient.callGraphQLService(instanceUrl, accessToken, null, query, variables);
    return response
        .getBodyAsJson()
        .getAsJsonObject("data")
        .getAsJsonPrimitive("completions")
        .getAsString();
  }

  /**
   * Sends a completions request to the Sourcegraph instance, and receives the response in a
   * streaming fashion.
   */
  public void streamCompletion(
      @NotNull CompletionsInput input,
      @NotNull CompletionsCallbacks cb,
      @NotNull Endpoint endpoint) {
    Gson gson =
        new GsonBuilder()
            .registerTypeAdapter(Speaker.class, new SpeakerLowercaseSerializer())
            .create();

    String body = gson.toJsonTree(input).getAsJsonObject().toString();

    SSEClient sseClient =
        new SSEClient(instanceUrl + endpoint.urlPath, accessToken, body, cb, endpoint);
    sseClient.start();
  }
}
