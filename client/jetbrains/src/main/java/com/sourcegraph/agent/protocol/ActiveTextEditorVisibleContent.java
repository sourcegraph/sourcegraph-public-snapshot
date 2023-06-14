package com.sourcegraph.agent.protocol;

import org.jetbrains.annotations.Nullable;

public class ActiveTextEditorVisibleContent {
  public String content;
  public String fileName;
  @Nullable public String repoName;

  public ActiveTextEditorVisibleContent setContent(String content) {
    this.content = content;
    return this;
  }

  public ActiveTextEditorVisibleContent setFileName(String fileName) {
    this.fileName = fileName;
    return this;
  }

  public ActiveTextEditorVisibleContent setRepoName(String repoName) {
    this.repoName = repoName;
    return this;
  }

  public ActiveTextEditorVisibleContent setRevision(String revision) {
    this.revision = revision;
    return this;
  }

  @Nullable public String revision;

  @Override
  public String toString() {
    return "ActiveTextEditorVisibleContent{"
        + "content='"
        + content
        + '\''
        + ", fileName='"
        + fileName
        + '\''
        + ", repoName='"
        + repoName
        + '\''
        + ", revision='"
        + revision
        + '\''
        + '}';
  }
}
