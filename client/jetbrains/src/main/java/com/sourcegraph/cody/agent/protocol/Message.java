package com.sourcegraph.cody.agent.protocol;

public class Message {
  public String speaker;
  public String text;

  @Override
  public String toString() {
    return "Message{" + "speaker='" + speaker + '\'' + ", text='" + text + '\'' + '}';
  }
}
