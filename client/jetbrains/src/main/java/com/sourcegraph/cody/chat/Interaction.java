package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.context.ContextMessage;
import java.util.List;
import java.util.Optional;

public class Interaction {
  private final ChatMessage humanMessage;
  private final List<ContextMessage> fullContext;
  private ChatMessage assistantMessage;

  public Interaction(ChatMessage humanMessage, List<ContextMessage> fullContext) {
    this.humanMessage = humanMessage;
    this.fullContext = fullContext;
    this.assistantMessage = ChatMessage.createAssistantMessage("");
  }

  public void setAssistantMessage(ChatMessage assistantMessage) {
    this.assistantMessage = assistantMessage;
  }

  public boolean hasContext() {
    return !this.fullContext.isEmpty();
  }

  public Optional<ChatMessage> getAssistantMessage() {
    return Optional.ofNullable(assistantMessage);
  }

  public ChatMessage getHumanMessage() {
    return humanMessage;
  }

  public List<ContextMessage> getFullContext() {
    return fullContext;
  }
}
