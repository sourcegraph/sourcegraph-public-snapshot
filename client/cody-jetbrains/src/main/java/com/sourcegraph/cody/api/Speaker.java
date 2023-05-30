package com.sourcegraph.cody.api;

public enum Speaker {
  HUMAN,
  ASSISTANT;

  public String prompt() {
    if (this == HUMAN) return "\n\nHuman:";
    return "\n\nAssistant:";
  }
}
