package com.sourcegraph.cody.api;

import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.vscode.CancellationToken;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ChatUpdaterCallbacks implements CompletionsCallbacks {
  private static final Logger logger = Logger.getInstance(ChatUpdaterCallbacks.class);
  private static final String STOP_SEQUENCE_REGEXP = "(H|Hu|Hum|Huma|Human|Human:)$";
  private static final Pattern stopSequencePattern = Pattern.compile(STOP_SEQUENCE_REGEXP);
  @NotNull private final UpdatableChat chat;
  @NotNull private final CancellationToken cancellationToken;
  @NotNull private final String prefix;
  private final AtomicBoolean gotFirstMessage = new AtomicBoolean(false);
  private final AtomicReference<String> lastMessage = new AtomicReference<>("");

  public ChatUpdaterCallbacks(
      @NotNull UpdatableChat chat,
      @NotNull CancellationToken cancellationToken,
      @NotNull String prefix) {
    this.chat = chat;
    this.cancellationToken = cancellationToken;
    this.prefix = prefix;
  }

  @Override
  public void onSubscribed() {
    logger.info("Subscribed to completions.");
  }

  @Override
  public void onData(@Nullable String data) {
    if (!cancellationToken.isCancelled())
      Optional.ofNullable(data)
          .ifPresent(
              d -> {
                String messageText = reformatBotMessage(d, prefix);
                // ward against streaming data coming in out of order
                if (messageText.length() > lastMessage.get().length()) {
                  lastMessage.set(messageText);
                  passMessageToChat(messageText);
                }
              });
  }

  private void passMessageToChat(@NotNull String messageText) {
    ChatMessage chatMessage = ChatMessage.createAssistantMessage(messageText);
    if (!gotFirstMessage.getAndSet(true)) chat.addMessageToChat(chatMessage);
    else chat.updateLastMessage(chatMessage);
  }

  @Override
  public void onError(@NotNull Throwable error) {
    if (!cancellationToken.isCancelled()) {
      String message = error.getMessage();
      chat.respondToErrorFromServer(message != null ? message : "");
      chat.finishMessageProcessing();
      logger.warn(error);
    }
  }

  @Override
  public void onComplete() {
    logger.info("Streaming completed.");
    if (!cancellationToken.isCancelled()) {
      chat.finishMessageProcessing();
    }
  }

  @Override
  public void onCancelled() {
    if (!cancellationToken.isCancelled()) {
      chat.finishMessageProcessing();
    }
  }

  private static @NotNull String reformatBotMessage(@NotNull String text, @NotNull String prefix) {
    String reformattedMessage = prefix + text.stripTrailing();
    Matcher stopSequenceMatcher = stopSequencePattern.matcher(reformattedMessage);
    return stopSequenceMatcher.find()
        ? reformattedMessage.substring(0, stopSequenceMatcher.start())
        : reformattedMessage;
  }
}
