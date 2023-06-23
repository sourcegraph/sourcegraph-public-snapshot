package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.CodyAgentClient;
import com.sourcegraph.cody.agent.CodyAgentServer;
import com.sourcegraph.cody.agent.protocol.ExecuteRecipeParams;
import com.sourcegraph.cody.api.ChatUpdaterCallbacks;
import com.sourcegraph.cody.api.CompletionsInput;
import com.sourcegraph.cody.api.CompletionsService;
import com.sourcegraph.cody.api.Message;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.vscode.CancellationToken;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicBoolean;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class Chat {
  private final @NotNull CompletionsService completionsService;

  /**
   * @param instanceUrl Like "https://sourcegraph.com/", with a slash at the end
   */
  public Chat(@NotNull String instanceUrl, @NotNull String accessToken) {
    completionsService = new CompletionsService(instanceUrl, accessToken);
  }

  public void sendMessageWithoutAgent(
      @NotNull List<Message> prompt, @Nullable String prefix, @NotNull UpdatableChat chat) {
    completionsService.streamCompletion(
        new CompletionsInput(prompt, 0.5f, null, 1000, -1, -1),
        new ChatUpdaterCallbacks(chat, prefix),
        CompletionsService.Endpoint.Stream,
        new CancellationToken());
  }

  public void sendMessageViaAgent(
      @NotNull CodyAgentClient client,
      @NotNull CompletableFuture<CodyAgentServer> codyAgentServer,
      @NotNull ChatMessage humanMessage,
      @NotNull UpdatableChat chat)
      throws ExecutionException, InterruptedException {
    final AtomicBoolean isFirstMessage = new AtomicBoolean(false);
    client.onChatUpdateMessageInProgress =
        (agentChatMessage) -> {
          if (agentChatMessage.text == null) {
            return;
          }
          ChatMessage chatMessage =
              new ChatMessage(
                  Speaker.ASSISTANT, agentChatMessage.text, agentChatMessage.displayText);
          if (isFirstMessage.compareAndSet(false, true)) {
            chat.addMessageToChat(chatMessage);
          } else {
            chat.updateLastMessage(chatMessage);
          }
        };

    codyAgentServer
        .thenAcceptAsync(
            server ->
                server.recipesExecute(
                    new ExecuteRecipeParams()
                        .setId("chat-question")
                        .setHumanChatInput(humanMessage.getText())),
            CodyAgent.executorService)
        .get();
    chat.finishMessageProcessing();
  }
}
