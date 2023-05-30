package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.api.Message;
import java.util.List;

public class Completion {
  public final String prefix;
  public final List<Message> messages;
  public final String content;
  public final String stopReason;

  public Completion(String prefix, List<Message> messages, String content, String stopReason) {
    this.prefix = prefix;
    this.messages = messages;
    this.content = content;
    this.stopReason = stopReason;
  }
}
