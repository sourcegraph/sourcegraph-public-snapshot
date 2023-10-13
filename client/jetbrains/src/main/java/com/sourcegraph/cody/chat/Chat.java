package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentClient;
import com.sourcegraph.cody.agent.CodyAgentServer;
import com.sourcegraph.cody.agent.protocol.ExecuteRecipeParams;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.context.ContextFile;
import com.sourcegraph.cody.context.ContextMessage;
import com.sourcegraph.cody.vscode.CancellationToken;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;

public class Chat {

  public void sendMessageViaAgent(
      @NotNull CodyAgentClient client,
      @NotNull CompletableFuture<CodyAgentServer> codyAgentServer,
      @NotNull ChatMessage humanMessage,
      @NotNull String recipeId,
      @NotNull UpdatableChat chat,
      @NotNull CancellationToken token)
      throws ExecutionException, InterruptedException {
    final AtomicBoolean isFirstMessage = new AtomicBoolean(false);
    client.onFinishedProcessing = chat::finishMessageProcessing;
    client.onChatUpdateMessageInProgress =
        (agentChatMessage) -> {
          if (agentChatMessage.text == null) {
            return;
          }

          ChatMessage chatMessage =
              new ChatMessage(
                  Speaker.ASSISTANT, agentChatMessage.text, agentChatMessage.displayText);
          if (isFirstMessage.compareAndSet(false, true)) {
            List<ContextMessage> contextMessages =
                agentChatMessage.actualContextFiles().stream()
                    .map(
                        contextFile ->
                            new ContextMessage(
                                Speaker.ASSISTANT,
                                agentChatMessage.text,
                                new ContextFile(
                                    contextFile.fileName,
                                    contextFile.repoName,
                                    contextFile.revision)))
                    .collect(Collectors.toList());
            chat.displayUsedContext(contextMessages);
            chat.addMessageToChat(chatMessage);
          } else {
            chat.updateLastMessage(chatMessage);
          }
        };

    codyAgentServer
        .thenAcceptAsync(
            server -> {
              try {
                CompletableFuture<Void> recipesExecuteFuture =
                    server.recipesExecute(
                        new ExecuteRecipeParams()
                            .setId(recipeId)
                            .setHumanChatInput(humanMessage.actualMessage()));
                token.onCancellationRequested(() -> recipesExecuteFuture.cancel(true));
              } catch (Exception ignored) {
                // Ignore bugs in the agent when executing recipes
              }
            },
            CodyAgent.executorService)
        .get();
  }
}
