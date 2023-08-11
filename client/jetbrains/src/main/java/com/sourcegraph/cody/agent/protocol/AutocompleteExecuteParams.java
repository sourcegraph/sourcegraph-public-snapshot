package com.sourcegraph.cody.agent.protocol;

public class AutocompleteExecuteParams {

  public String filePath;
  public Position position;
  public AutocompleteContext context;

  public AutocompleteExecuteParams setFilePath(String filePath) {
    this.filePath = filePath;
    return this;
  }

  public AutocompleteExecuteParams setPosition(Position position) {
    this.position = position;
    return this;
  }

  public AutocompleteExecuteParams setContext(AutocompleteContext context) {
    this.context = context;
    return this;
  }
}
