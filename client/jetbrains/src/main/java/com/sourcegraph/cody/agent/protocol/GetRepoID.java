package com.sourcegraph.cody.agent.protocol;

public class GetRepoID {
  public final String repoName;

  public GetRepoID(String repoName) {
    this.repoName = repoName;
  }
}
