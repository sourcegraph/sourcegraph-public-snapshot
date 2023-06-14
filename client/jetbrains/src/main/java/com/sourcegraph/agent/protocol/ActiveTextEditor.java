package com.sourcegraph.agent.protocol;

import org.jetbrains.annotations.Nullable;

public class ActiveTextEditor {

  public String content;
  public String filePath;
  @Nullable public String repoName;
  @Nullable public String revision;

  public ActiveTextEditor setContent(String content) {
    this.content = content;
    return this;
  }

  public ActiveTextEditor setFilePath(String filePath) {
    this.filePath = filePath;
    return this;
  }

  public ActiveTextEditor setRepoName(String repoName) {
    this.repoName = repoName;
    return this;
  }

  public ActiveTextEditor setRevision(String revision) {
    this.revision = revision;
    return this;
  }
}
