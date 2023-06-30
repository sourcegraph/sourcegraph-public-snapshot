package com.sourcegraph.cody.context.embeddings;

import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonPrimitive;
import com.google.gson.JsonSyntaxException;
import com.sourcegraph.api.GraphQlClient;
import com.sourcegraph.api.GraphQlResponse;
import com.sourcegraph.cody.context.ContextMessage;
import com.sourcegraph.cody.prompts.Prompter;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class EmbeddingsSearcher {
  private final @NotNull String instanceUrl;
  private final @NotNull String accessToken;
  private final @NotNull String customRequestHeaders;

  public EmbeddingsSearcher(
      @NotNull String instanceUrl,
      @NotNull String accessToken,
      @NotNull String customRequestHeaders) {
    this.instanceUrl = instanceUrl;
    this.accessToken = accessToken;
    this.customRequestHeaders = customRequestHeaders;
  }

  /**
   * @param repoName Like "github.com/sourcegraph/cody"
   */
  public @NotNull List<ContextMessage> getContextMessages(
      @NotNull String repoName, @NotNull String query, int codeResultCount, int textResultCount)
      throws IOException {
    // Get repo ID
    String repoId;
    repoId = getRepoIdIfEmbeddingExists(repoName);
    if (repoId == null) { // Embedding does not exist
      return new ArrayList<>();
    }

    // Run embeddings search
    EmbeddingsSearchResults results = search(repoId, query, codeResultCount, textResultCount);

    // Concat results.getCodeResults() and results.getTextResults() into a single list
    List<EmbeddingsSearchResult> allResults = new ArrayList<>();
    allResults.addAll(results.getCodeResults());
    allResults.addAll(results.getTextResults());

    // Group results by file
    List<GroupedResults> groupedResults = ResultsGrouper.groupResultsByFile(allResults);

    // Reverse results so that they appear in ascending order of importance (least -> most)
    Collections.reverse(groupedResults);

    // Get context messages
    List<ContextMessage> messages = new ArrayList<>();
    for (GroupedResults group : groupedResults) {
      for (String snippet : group.getSnippets()) {
        String contextPrompt = Prompter.getContextPrompt(group.getFile().getFileName(), snippet);
        messages.add(ContextMessage.createHumanMessage(contextPrompt, group.getFile()));
        messages.add(ContextMessage.createDefaultAssistantMessage());
      }
    }
    return messages;
  }

  private EmbeddingsSearchResults search(
      @NotNull String repoId, @NotNull String query, int codeResultsCount, int textResultsCount)
      throws IOException {
    // Prepare GraphQL query
    String graphQlQuery =
        "query LegacyEmbeddingsSearch($repo: ID!, $query: String!, $codeResultsCount: Int!, $textResultsCount: Int!) {\n"
            + "    embeddingsSearch(repo: $repo, query: $query, codeResultsCount: $codeResultsCount, textResultsCount: $textResultsCount) {\n"
            + "        codeResults {\n"
            + "            fileName\n"
            + "            startLine\n"
            + "            endLine\n"
            + "            content\n"
            + "        }\n"
            + "        textResults {\n"
            + "            fileName\n"
            + "            startLine\n"
            + "            endLine\n"
            + "            content\n"
            + "        }\n"
            + "    }\n"
            + "}";
    JsonObject variables = new JsonObject();
    variables.add("repo", new JsonPrimitive(repoId));
    variables.add("query", new JsonPrimitive(query));
    variables.add("codeResultsCount", new JsonPrimitive(codeResultsCount));
    variables.add("textResultsCount", new JsonPrimitive(textResultsCount));

    // Call GraphQL service
    GraphQlResponse response =
        GraphQlClient.callGraphQLService(
            instanceUrl, accessToken, customRequestHeaders, graphQlQuery, variables);

    // Parse response
    if (response.getStatusCode() != 200) {
      throw new IOException("GraphQL request failed with status code " + response.getStatusCode());
    } else {
      try {
        JsonObject body = response.getBodyAsJson();
        if (body.has("errors")) {
          throw new IOException("GraphQL request failed with errors: " + body.get("errors"));
        }
        JsonObject data = body.getAsJsonObject("data");
        if (data == null) {
          throw new IOException("GraphQL response is missing data field");
        }
        JsonObject embeddingsSearch = data.getAsJsonObject("embeddingsSearch");
        if (embeddingsSearch == null) {
          throw new IOException("GraphQL response is missing data.embeddingsSearch field");
        }

        ArrayList<EmbeddingsSearchResult> codeResults =
            convertRawResultsToSearchResults(embeddingsSearch.getAsJsonArray("codeResults"));
        ArrayList<EmbeddingsSearchResult> textResults =
            convertRawResultsToSearchResults(embeddingsSearch.getAsJsonArray("textResults"));
        return new EmbeddingsSearchResults(codeResults, textResults);

      } catch (JsonSyntaxException e) {
        throw new IOException("GraphQL response is not valid JSON", e);
      }
    }
  }

  /**
   * Converts raw results from a GraphQL response to a list of EmbeddingsSearchResult objects. This
   * works for both code and text results.
   */
  private static @NotNull ArrayList<EmbeddingsSearchResult> convertRawResultsToSearchResults(
      @Nullable JsonArray rawResults) {
    if (rawResults == null) {
      return new ArrayList<>();
    }
    ArrayList<EmbeddingsSearchResult> results = new ArrayList<>();
    for (JsonElement result : rawResults) {
      JsonPrimitive repoName = ((JsonObject) result).getAsJsonPrimitive("repoName");
      JsonPrimitive revision = ((JsonObject) result).getAsJsonPrimitive("revision");
      String fileName = ((JsonObject) result).getAsJsonPrimitive("fileName").getAsString();
      int startLine = ((JsonObject) result).getAsJsonPrimitive("startLine").getAsInt();
      int endLine = ((JsonObject) result).getAsJsonPrimitive("endLine").getAsInt();
      String content = ((JsonObject) result).getAsJsonPrimitive("content").getAsString();
      results.add(
          new EmbeddingsSearchResult(
              repoName != null ? repoName.toString() : null,
              revision != null ? revision.toString() : null,
              fileName,
              startLine,
              endLine,
              content));
    }
    return results;
  }

  /**
   * Returns the repository ID if the repository exists and has an embedding, or null otherwise.
   *
   * @param repoName Like "github.com/sourcegraph/cody"
   * @return base64-encoded repoID like "UmVwb3NpdG9yeTozN1gwOTI1MA=="
   * @throws IOException Thrown if we can't reach the server.
   */
  private @Nullable String getRepoIdIfEmbeddingExists(String repoName) throws IOException {
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
