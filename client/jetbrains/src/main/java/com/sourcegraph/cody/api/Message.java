package com.sourcegraph.cody.api;

import org.jetbrains.annotations.NotNull;

public class Message {
  protected final @NotNull Speaker speaker;
  protected final @NotNull String text;

  public Message(@NotNull Speaker speaker, @NotNull String text) {
    this.speaker = speaker;
    this.text = text;
  }

  public @NotNull Speaker getSpeaker() {
    return speaker;
  }

  public @NotNull String getText() {
    return text;
  }

  public @NotNull String prompt() {
    return speaker.prompt() + (text.isEmpty() ? "" : " " + text);
  }

  @Override
  public @NotNull String toString() {
    return String.format("Message { speaker=%s, text='%s'}", speaker, text);
  }
}
