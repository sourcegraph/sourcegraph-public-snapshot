package com.sourcegraph.cody.agent;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.Editor;
import com.sourcegraph.cody.agent.protocol.ChatMessage;
import com.sourcegraph.cody.agent.protocol.DebugMessage;
import java.util.concurrent.CompletableFuture;
import java.util.function.Consumer;
import java.util.function.Supplier;
import org.eclipse.lsp4j.jsonrpc.services.JsonNotification;
import org.jetbrains.annotations.Nullable;

/** Implementation of the client part of the Cody agent protocol. */
@SuppressWarnings("unused")
public class CodyAgentClient {

  private static final Logger logger = Logger.getInstance(CodyAgentClient.class);
  @Nullable public CodyAgentServer server;
  @Nullable public CodyAgentDocuments documents;
  // Callback that is invoked when the agent sends a "chat/updateMessageInProgress" notification.
  @Nullable public Consumer<ChatMessage> onChatUpdateMessageInProgress;
  @Nullable public Editor editor;

  /**
   * Helper to run client request/notification handlers on the IntelliJ event thread. Use this
   * helper for handlers that require access to the IntelliJ editor, for example to read the text
   * contents of the open editor.
   */
  private <T> CompletableFuture<T> onEventThread(Supplier<T> handler) {
    CompletableFuture<T> result = new CompletableFuture<>();
    ApplicationManager.getApplication()
        .invokeLater(
            () -> {
              try {
                result.complete(handler.get());
              } catch (Exception e) {
                result.completeExceptionally(e);
              }
            });
    return result;
  }

  // =============
  // Notifications
  // =============

  @JsonNotification("chat/updateMessageInProgress")
  public void chatUpdateMessageInProgress(ChatMessage params) {
    if (onChatUpdateMessageInProgress != null && params != null) {
      ApplicationManager.getApplication()
          .invokeLater(() -> onChatUpdateMessageInProgress.accept(params));
    }
  }

  @JsonNotification("debug/message")
  public void debugMessage(DebugMessage msg) {
    logger.warn(String.format("%s: %s", msg.channel, msg.message));
  }
}
