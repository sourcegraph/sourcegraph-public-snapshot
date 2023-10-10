package com.sourcegraph.cody.agent.protocol;

import org.jetbrains.annotations.Nullable;

public class Message {
  @Nullable private String speaker;
  @Nullable public String text;

  @Override
  public String toString() {
    return "Message{" + "speaker='" + speaker + '\'' + ", text='" + text + '\'' + '}';
  }
}
