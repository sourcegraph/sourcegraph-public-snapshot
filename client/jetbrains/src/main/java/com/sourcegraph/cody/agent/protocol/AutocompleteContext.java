package com.sourcegraph.cody.agent.protocol;

public class AutocompleteContext {
  public String triggerKind;

  public AutocompleteContext withInvokeTriggerKind() {
    this.triggerKind = "invoke";
    return this;
  }

  public AutocompleteContext withAutomaticTriggerKind() {
    this.triggerKind = "automatic";
    return this;
  }
}
