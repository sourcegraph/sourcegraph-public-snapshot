package com.sourcegraph.cody.agent.protocol;

import org.jetbrains.annotations.Nullable;

public class AutocompleteContext {
  @Nullable public String triggerKind;

  public AutocompleteContext withInvokeTriggerKind() {
    this.triggerKind = "invoke";
    return this;
  }

  public AutocompleteContext withAutomaticTriggerKind() {
    this.triggerKind = "automatic";
    return this;
  }
}
