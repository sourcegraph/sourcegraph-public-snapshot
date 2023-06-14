package com.sourcegraph.agent.protocol;

public class InteractionMessage extends Message {
  public String displayText;
  public String prefix;

  @Override
  public String toString() {
    return "InteractionMessage{"
        + "displayText='"
        + displayText
        + '\''
        + ", prefix='"
        + prefix
        + '\''
        + ", speaker='"
        + speaker
        + '\''
        + ", text='"
        + text
        + '\''
        + '}';
  }
}
