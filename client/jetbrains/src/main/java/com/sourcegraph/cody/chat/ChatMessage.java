package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Speaker;
import org.jetbrains.annotations.NotNull;

public class ChatMessage extends Message {
  private final @NotNull String displayText;

  ChatMessage(@NotNull Speaker speaker, @NotNull String text, @NotNull String displayText) {
    super(speaker, text);
    this.displayText = displayText;
  }

  public static @NotNull ChatMessage createAssistantMessage(@NotNull String text) {
    return new ChatMessage(Speaker.ASSISTANT, text, text);
  }

  public static @NotNull ChatMessage createHumanMessage(
      @NotNull String prompt, @NotNull String displayText) {
    return new ChatMessage(Speaker.HUMAN, prompt, displayText);
  }

  /* This is considered a markdown formatted string */
  @NotNull
  public String getDisplayText() {
    return displayText;
  }
}
