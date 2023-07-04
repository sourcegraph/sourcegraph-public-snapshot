package com.sourcegraph.cody.context.embeddings;

import com.google.gson.JsonObject;
import com.google.gson.JsonPrimitive;
import com.google.gson.JsonSyntaxException;
import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.api.GraphQlClient;
import com.sourcegraph.api.GraphQlResponse;
import java.io.IOException;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class EmbeddingsStatusLoader {

  private static final Logger logger = Logger.getInstance(EmbeddingsStatusLoader.class);

  private final @NotNull String instanceUrl;
  private final @NotNull String accessToken;
  private final @NotNull String customRequestHeaders;

  public EmbeddingsStatusLoader(
      @NotNull String instanceUrl,
      @NotNull String accessToken,
      @NotNull String customRequestHeaders) {
    this.instanceUrl = instanceUrl;
    this.accessToken = accessToken;
    this.customRequestHeaders = customRequestHeaders;
  }

  public @Nullable String getRepoId(@NotNull String repoName) {
    try {
      return getRepoIdIfEmbeddingExists(repoName);
    } catch (IOException e) {
      logger.warn("Unable to load repo id for " + repoName, e);
      return null;
    }
  }

  /**
   * Returns the repository ID if the repository exists and has an embedding, or null otherwise.
   *
   * @param repoName Like "github.com/sourcegraph/cody"
   * @return base64-encoded repoID like "UmVwb3NpdG9yeTozN1gwOTI1MA=="
   * @throws IOException Thrown if we can't reach the server.
   */
  public @Nullable String getRepoIdIfEmbeddingExists(@NotNull String repoName) throws IOException {
    String query =
        "query Repository($name: String!) {\n"
            + "    repository(name: $name) {\n"
            + "        id\n"
            + "        embeddingExists\n"
            + "    }\n"
            + "}";
    JsonObject variables = new JsonObject();
    variables.add("name", new JsonPrimitive(repoName));
    GraphQlResponse response =
        GraphQlClient.callGraphQLService(
            instanceUrl, accessToken, customRequestHeaders, query, variables);
    if (response.getStatusCode() != 200) {
      throw new IOException("GraphQL request failed with status code " + response.getStatusCode());
    } else {
      try {
        JsonObject body = response.getBodyAsJson();
        JsonObject data = body.getAsJsonObject("data");
        if (data == null || data.get("repository").isJsonNull()) { // Embedding does not exist
          return null;
        }
        JsonObject repository = data.getAsJsonObject("repository");
        boolean embeddingExists = repository.getAsJsonPrimitive("embeddingExists").getAsBoolean();
        if (embeddingExists) {
          return repository.getAsJsonPrimitive("id").getAsString();
        } else {
          return null;
        }
      } catch (JsonSyntaxException | ClassCastException e) {
        throw new IOException("GraphQL response is not valid JSON", e);
      }
    }
  }
}
