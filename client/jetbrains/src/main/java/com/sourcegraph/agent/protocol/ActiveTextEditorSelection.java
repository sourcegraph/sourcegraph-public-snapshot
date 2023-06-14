package com.sourcegraph.agent.protocol;

public class ActiveTextEditorSelection {
  public String fileName;
  public String repoName;
  public String revision;
  public String precedingText;
  public String selectedText;

  public ActiveTextEditorSelection setFileName(String fileName) {
    this.fileName = fileName;
    return this;
  }

  public ActiveTextEditorSelection setRepoName(String repoName) {
    this.repoName = repoName;
    return this;
  }

  public ActiveTextEditorSelection setRevision(String revision) {
    this.revision = revision;
    return this;
  }

  public ActiveTextEditorSelection setPrecedingText(String precedingText) {
    this.precedingText = precedingText;
    return this;
  }

  public ActiveTextEditorSelection setSelectedText(String selectedText) {
    this.selectedText = selectedText;
    return this;
  }

  public ActiveTextEditorSelection setFollowingText(String followingText) {
    this.followingText = followingText;
    return this;
  }

  public String followingText;
}
