package com.sourcegraph.agent.protocol;

public class ExecuteRecipeParams {

  public String id;
  public String humanChatInput;
  public StaticRecipeContext context;

  public ExecuteRecipeParams setId(String id) {
    this.id = id;
    return this;
  }

  public ExecuteRecipeParams setHumanChatInput(String humanChatInput) {
    this.humanChatInput = humanChatInput;
    return this;
  }

  public ExecuteRecipeParams setContext(StaticRecipeContext context) {
    this.context = context;
    return this;
  }
}
