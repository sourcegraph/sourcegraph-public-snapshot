package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Speaker;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ChatMessage extends Message {
  private final @Nullable String displayText;

  public ChatMessage(@NotNull Speaker speaker, @NotNull String text, @Nullable String displayText) {
    super(speaker, text);
    this.displayText = displayText;
  }

  public ChatMessage(@NotNull Speaker speaker, @NotNull String text) {
    this(speaker, text, null);
  }

  @NotNull
  public String actualMessage() {
    return displayText != null ? displayText : text != null ? text : "";
  }
}
