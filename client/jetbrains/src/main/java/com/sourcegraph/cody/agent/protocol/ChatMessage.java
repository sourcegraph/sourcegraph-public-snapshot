package com.sourcegraph.cody.agent.protocol;

import java.util.List;
import org.jetbrains.annotations.Nullable;

// TODO: consolidate with the other ChatMessage. This duplication exists because the other
// ChatMessage uses an enum for Message.speaker that doesn't decode nicely with Gson.
public class ChatMessage extends Message {
  @Nullable public String displayText;
  @Nullable private List<ContextFile> contextFiles;

  public ChatMessage(@Nullable List<ContextFile> contextFiles) {
    this.contextFiles = contextFiles;
  }

  public List<ContextFile> actualContextFiles() {
    return contextFiles != null ? contextFiles : List.of();
  }

  @Override
  public String toString() {
    return "ChatMessage{"
        + "displayText='"
        + displayText
        + '\''
        + ", contextFiles="
        + contextFiles
        + ", text='"
        + text
        + '\''
        + '}';
  }
}
