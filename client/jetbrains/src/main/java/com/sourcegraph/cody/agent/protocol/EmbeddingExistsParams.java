package com.sourcegraph.cody.agent.protocol;

public class EmbeddingExistsParams {
  public final String repoName;

  public EmbeddingExistsParams(String repoName) {
    this.repoName = repoName;
  }
}
