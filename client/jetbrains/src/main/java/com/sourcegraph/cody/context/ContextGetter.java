package com.sourcegraph.cody.context;

import com.sourcegraph.cody.context.embeddings.EmbeddingsSearcher;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import org.jetbrains.annotations.NotNull;

public class ContextGetter {
  private final @NotNull String repoName;
  private final @NotNull EmbeddingsSearcher embeddingsSearcher;

  /**
   * @param repoName    Like "github.com/sourcegraph/cody"
   * @param instanceUrl Like "https://sourcegraph.com/", with a slash at the end
   */
  public ContextGetter(@NotNull String repoName, @NotNull String instanceUrl,
      @NotNull String accessToken, @NotNull String customRequestHeaders) {
    this.repoName = repoName;
    embeddingsSearcher = new EmbeddingsSearcher(instanceUrl, accessToken, customRequestHeaders);
  }

  public @NotNull List<ContextMessage> getContextMessages(
      @NotNull String query, int codeResultCount, int textResultCount, boolean useEmbeddings)
      throws IOException {
    if (useEmbeddings) {
      return embeddingsSearcher.getContextMessages(
          repoName, query, codeResultCount, textResultCount);
    } else {
      // TODO: Add keyword search if embeddings are not available
      // return KeywordSearcher.getContextMessages(query, codeResultCount, textResultCount);
      return new ArrayList<>();
    }
  }
}
