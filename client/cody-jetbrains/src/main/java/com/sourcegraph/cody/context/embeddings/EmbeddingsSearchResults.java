package com.sourcegraph.cody.context.embeddings;

import java.util.List;
import org.jetbrains.annotations.NotNull;

public class EmbeddingsSearchResults {
  private @NotNull List<EmbeddingsSearchResult> codeResults;
  private @NotNull List<EmbeddingsSearchResult> textResults;

  public EmbeddingsSearchResults(
      @NotNull List<EmbeddingsSearchResult> codeResults,
      @NotNull List<EmbeddingsSearchResult> textResults) {
    this.codeResults = codeResults;
    this.textResults = textResults;
  }

  // Getters and setters
  public @NotNull List<EmbeddingsSearchResult> getCodeResults() {
    return codeResults;
  }

  public void setCodeResults(@NotNull List<EmbeddingsSearchResult> codeResults) {
    this.codeResults = codeResults;
  }

  public @NotNull List<EmbeddingsSearchResult> getTextResults() {
    return textResults;
  }

  public void setTextResults(@NotNull List<EmbeddingsSearchResult> textResults) {
    this.textResults = textResults;
  }
}
