package com.sourcegraph.cody.agent.protocol;

import java.util.List;

// TODO: consolidate with the other ChatMessage. This duplication exists because the other
// ChatMessage uses an enum for Message.speaker that doesn't decode nicely with Gson.
public class ChatMessage extends Message {
  public String displayText;
  public List<ContextFile> contextFiles;

  @Override
  public String toString() {
    return "ChatMessage{"
        + "displayText='"
        + displayText
        + '\''
        + ", contextFiles="
        + contextFiles
        + ", speaker='"
        + speaker
        + '\''
        + ", text='"
        + text
        + '\''
        + '}';
  }
}
