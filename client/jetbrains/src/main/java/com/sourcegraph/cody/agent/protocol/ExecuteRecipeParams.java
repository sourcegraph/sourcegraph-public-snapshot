package com.sourcegraph.cody.agent.protocol;

public class ExecuteRecipeParams {

  public String id;
  public String humanChatInput;

  public ExecuteRecipeParams setId(String id) {
    this.id = id;
    return this;
  }

  public ExecuteRecipeParams setHumanChatInput(String humanChatInput) {
    this.humanChatInput = humanChatInput;
    return this;
  }
}
