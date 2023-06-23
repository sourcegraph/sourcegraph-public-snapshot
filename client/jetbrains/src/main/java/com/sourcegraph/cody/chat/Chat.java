package com.sourcegraph.cody.chat;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.protocol.ExecuteRecipeParams;
import com.sourcegraph.cody.api.*;
import com.sourcegraph.cody.prompts.Preamble;
import com.sourcegraph.cody.vscode.CancellationToken;
import java.util.ArrayList;
import java.util.List;
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
    if (CodyAgent.isConnected(project)) {
      sendMessageViaAgent(project, humanMessage, chat);
    } else {
      sendMessageWithoutAgent(humanMessage, prefix, chat);
    }
  }

  private void sendMessageWithoutAgent(
      @NotNull ChatMessage humanMessage, @Nullable String prefix, @NotNull UpdatableChat chat) {
    // TODO: Usethe context getting logic from VS Code
    List<Message> preamble = Preamble.getPreamble(codebase);
    var codeContext = "";
    if (humanMessage.getContextFiles().size() == 0) {
      codeContext = "I have no file open in the editor right now.";
    } else {
      codeContext = "Here is my current file\n" + humanMessage.getContextFiles().get(0);
    }

    var input = new CompletionsInput(new ArrayList<>(), 0.5f, null, 1000, -1, -1);
    input.addMessages(preamble);
    input.addMessage(Speaker.HUMAN, codeContext);
    input.addMessage(Speaker.ASSISTANT, "Ok.");
    input.addMessage(Speaker.HUMAN, humanMessage.getText());
    input.addMessage(Speaker.ASSISTANT, "");

    input.messages.forEach(System.out::println);

    // ConfigUtil.getAccessToken(project) TODO: Get the access token from the plugin config
    // TODO: Don't create this each time
    completionsService.streamCompletion(
        input,
        new ChatUpdaterCallbacks(chat, prefix),
        CompletionsService.Endpoint.Stream,
        new CancellationToken());
  }

  private void sendMessageViaAgent(
      @NotNull Project project, @NotNull ChatMessage humanMessage, @NotNull UpdatableChat chat)
      throws ExecutionException, InterruptedException {
    final AtomicBoolean isFirstMessage = new AtomicBoolean(false);
    CodyAgent.getClient(project).onChatUpdateMessageInProgress =
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

    CodyAgent.getInitializedServer(project)
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
