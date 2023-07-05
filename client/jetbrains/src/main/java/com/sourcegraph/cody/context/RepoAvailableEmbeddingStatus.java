package com.sourcegraph.cody.context;

import org.jetbrains.annotations.NotNull;

public abstract class RepoAvailableEmbeddingStatus implements EmbeddingStatus {
  private final String simpleRepositoryName;

  protected RepoAvailableEmbeddingStatus(String fullRepositoryName) {
    this.simpleRepositoryName = getRepositoryNameAfterLastSlash(fullRepositoryName);
  }

  @NotNull
  private static String getRepositoryNameAfterLastSlash(String fullRepositoryName) {
    int indexOfLastSlash = fullRepositoryName.lastIndexOf('/');
    String repoNameWithoutTrailingSlash =
        indexOfLastSlash == fullRepositoryName.length() - 1
            ? fullRepositoryName.substring(0, indexOfLastSlash)
            : fullRepositoryName;
    indexOfLastSlash = repoNameWithoutTrailingSlash.lastIndexOf('/');
    return indexOfLastSlash != -1 && indexOfLastSlash != repoNameWithoutTrailingSlash.length() - 1
        ? repoNameWithoutTrailingSlash.substring(indexOfLastSlash + 1)
        : repoNameWithoutTrailingSlash;
  }

  @Override
  public @NotNull String getMainText() {
    return simpleRepositoryName;
  }
}
