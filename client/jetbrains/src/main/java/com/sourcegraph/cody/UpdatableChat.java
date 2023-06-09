package com.sourcegraph.cody;

import com.sourcegraph.cody.chat.ChatMessage;
import org.jetbrains.annotations.NotNull;

public interface UpdatableChat {
  void addMessageToChat(@NotNull ChatMessage message);

  void respondToMessage(@NotNull ChatMessage message);

  void updateLastMessage(@NotNull ChatMessage message);

  void finishMessageProcessing();

  void resetConversation();

  void activateChatTab();
}
