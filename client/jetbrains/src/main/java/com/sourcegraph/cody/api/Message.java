package com.sourcegraph.cody.api;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class Message {
  public final @NotNull Speaker speaker;
  protected final @Nullable String text;

  public Message(@NotNull Speaker speaker, @Nullable String text) {
    this.speaker = speaker;
    this.text = text;
  }
}
