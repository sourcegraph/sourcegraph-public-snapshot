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
  private final @NotNull CompletionsService completionsService;

  /**
   * @param instanceUrl Like "https://sourcegraph.com/", with a slash at the end
   */
  public Chat(@NotNull String instanceUrl, @NotNull String accessToken) {
    completionsService = new CompletionsService(instanceUrl, accessToken);
  }

  public void sendMessageWithoutAgent(
      @NotNull List<Message> prompt,
      @NotNull String prefix,
      @NotNull UpdatableChat chat,
      @NotNull CancellationToken cancellationToken) {
    completionsService.streamCompletion(
        new CompletionsInput(prompt, 0.5f, null, 1000, -1, -1),
        new ChatUpdaterCallbacks(chat, cancellationToken, prefix),
        CompletionsService.Endpoint.Stream);
  }

  public void sendMessageViaAgent(
      @NotNull CodyAgentClient client,
      @NotNull CompletableFuture<CodyAgentServer> codyAgentServer,
      @NotNull ChatMessage humanMessage,
      @NotNull UpdatableChat chat,
      @NotNull CancellationToken cancellationToken)
      throws ExecutionException, InterruptedException {
    final AtomicBoolean isFirstMessage = new AtomicBoolean(false);
    client.onChatUpdateMessageInProgress =
        (agentChatMessage) -> {
          if (agentChatMessage.text == null || cancellationToken.isCancelled()) {
            return;
          }

          ChatMessage chatMessage =
              new ChatMessage(
                  Speaker.ASSISTANT, agentChatMessage.text, agentChatMessage.displayText);
          if (isFirstMessage.compareAndSet(false, true)) {
            List<ContextMessage> contextMessages =
                agentChatMessage.contextFiles.stream()
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
            server ->
                server.recipesExecute(
                    new ExecuteRecipeParams()
                        .setId("chat-question")
                        .setHumanChatInput(humanMessage.getText())),
            CodyAgent.executorService)
        .get();
    // TODO we need to move this finishMessageProcessing to be executed when the whole message
    // processing is finished to make "stop generating" works. Ideally we need a signal from agent
    // that it finished processing the message so we can call this method.
    chat.finishMessageProcessing();
  }
}
