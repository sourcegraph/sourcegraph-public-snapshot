package com.sourcegraph.cody;

import com.sourcegraph.cody.chat.ChatMessage;
import org.jetbrains.annotations.NotNull;

public interface UpdatableChat {
  void addMessage(@NotNull ChatMessage message);

  void updateLastMessage(@NotNull ChatMessage message);

  void finishMessageProcessing();

  void resetConversation();
}
