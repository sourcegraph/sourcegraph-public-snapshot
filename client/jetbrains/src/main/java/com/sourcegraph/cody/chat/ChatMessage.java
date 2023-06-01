package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Speaker;
import java.util.ArrayList;
import org.jetbrains.annotations.NotNull;

public class ChatMessage extends Message {
  private final @NotNull String displayText;
  private final @NotNull ArrayList<String> contextFiles;

  private ChatMessage(
      @NotNull Speaker speaker,
      @NotNull String text,
      @NotNull String displayText,
      @NotNull ArrayList<String> contextFiles) {
    super(speaker, text);
    this.displayText = displayText;
    this.contextFiles = contextFiles;
  }

  public static @NotNull ChatMessage createAssistantMessage(@NotNull String text) {
    return new ChatMessage(Speaker.ASSISTANT, text, text, new ArrayList<>());
  }

  public static @NotNull ChatMessage createHumanMessage(
      @NotNull String text, @NotNull ArrayList<String> contextFiles) {
    return new ChatMessage(Speaker.HUMAN, text, text, contextFiles);
  }

  @NotNull
  public String getDisplayText() {
    return displayText;
  }

  @NotNull
  public ArrayList<String> getContextFiles() {
    return contextFiles;
  }
}
