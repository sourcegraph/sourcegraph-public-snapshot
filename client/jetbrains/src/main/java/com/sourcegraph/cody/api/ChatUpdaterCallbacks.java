package com.sourcegraph.cody.api;

import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.vscode.CancellationToken;
import java.util.Optional;
import java.util.concurrent.*;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ChatUpdaterCallbacks implements CompletionsCallbacks {
  private static final Logger logger = Logger.getInstance(ChatUpdaterCallbacks.class);
  private static final ScheduledExecutorService scheduler =
      Executors.newSingleThreadScheduledExecutor();
  private static final String STOP_SEQUENCE_REGEXP = "(H|Hu|Hum|Huma|Human|Human:)$";
  private static final Pattern stopSequencePattern = Pattern.compile(STOP_SEQUENCE_REGEXP);
  @NotNull private final UpdatableChat chat;
  @NotNull private final CancellationToken cancellationToken;
  @NotNull private final String prefix;
  private final AtomicBoolean gotFirstMessage = new AtomicBoolean(false);
  private final AtomicBoolean isCompleted = new AtomicBoolean(false);
  private final AtomicReference<String> lastMessage = new AtomicReference<>("");
  private final BlockingQueue<String> queue = new LinkedBlockingQueue<>();

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
    if (!cancellationToken.isCancelled()) {
      scheduler.schedule(
          () -> {
            while (!Thread.currentThread().isInterrupted()
                && !cancellationToken.isCancelled()
                && (!queue.isEmpty() || !isCompleted.get())) {
              try {
                Optional.ofNullable(queue.poll(5, TimeUnit.MILLISECONDS))
                    .ifPresent(this::passMessageToChat);
              } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
              }
            }
            chat.finishMessageProcessing();
          },
          20,
          TimeUnit.MILLISECONDS);
    }
  }

  @Override
  public void onData(@Nullable String data) {
    if (!cancellationToken.isCancelled())
      Optional.ofNullable(data)
          .ifPresent(
              d -> {
                String newMessage = reformatBotMessage(d, prefix);
                String previousMessage = lastMessage.get();
                String toAppend =
                    newMessage.startsWith(previousMessage)
                        ? StringUtils.stripStart(newMessage, previousMessage)
                        : newMessage;
                for (int i = 1; i <= toAppend.length(); i++) {
                  String messageToPass = previousMessage + toAppend.substring(0, i);
                  if (messageToPass.length() > lastMessage.get().length())
                    queue.offer(messageToPass);
                }
              });
  }

  private void passMessageToChat(@NotNull String messageText) {
    lastMessage.set(messageText);
    ChatMessage chatMessage = ChatMessage.createAssistantMessage(messageText);
    if (!gotFirstMessage.getAndSet(true)) chat.addMessageToChat(chatMessage);
    else chat.updateLastMessage(chatMessage);
  }

  @Override
  public void onError(@NotNull Throwable error) {
    isCompleted.set(true);
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
    isCompleted.set(true);
  }

  @Override
  public void onCancelled() {
    isCompleted.set(true);
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
