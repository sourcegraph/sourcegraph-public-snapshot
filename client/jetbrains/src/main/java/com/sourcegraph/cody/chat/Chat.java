package com.sourcegraph.cody.chat;

import com.intellij.openapi.project.Project;
import com.sourcegraph.agent.*;
import com.sourcegraph.agent.protocol.ExecuteRecipeParams;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.api.*;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class Chat {
  private final @Nullable String codebase;
  private final @NotNull CompletionsService completionsService;

  public Chat(@Nullable String codebase, @NotNull String instanceUrl, @NotNull String accessToken) {
    this.codebase = codebase;
    completionsService = new CompletionsService(instanceUrl, accessToken);
  }

  public void sendMessage(
      @NotNull Project project,
      @NotNull ChatMessage humanMessage,
      @Nullable String prefix,
      @NotNull UpdatableChat chat)
      throws ExecutionException, InterruptedException {
    final AtomicBoolean isFirstMessage = new AtomicBoolean(false);
    CodyAgent.getClient().onChatUpdateMessageInProgress =
        (agentChatMessage) -> {
          if (agentChatMessage.text == null) {
            return;
          }
          ChatMessage chatMessage =
              new ChatMessage(
                  Speaker.ASSISTANT,
                  agentChatMessage.text,
                  agentChatMessage.displayText,
                  agentChatMessage.contextFiles.stream()
                      .map(file -> file.fileName)
                      .collect(Collectors.toList()));
          if (isFirstMessage.compareAndSet(false, true)) {
            chat.addMessageToChat(chatMessage);
          } else {
            chat.updateLastMessage(chatMessage);
          }
        };

    if (!CodyAgent.isConnected()) {
      // TODO: better error handling.
      chat.addMessageToChat(
          ChatMessage.createAssistantMessage(
              "TODO: show helpful error message explaining how to fix the Agent connection. "
                  + "The chat should probably be disabled when the agent is not connected."));
      return;
    }

    CodyAgent.getServer()
        .recipesExecute(
            new ExecuteRecipeParams()
                .setId("chat-question")
                .setHumanChatInput(humanMessage.getText()))
        .get();
    chat.finishMessageProcessing();
  }
}
