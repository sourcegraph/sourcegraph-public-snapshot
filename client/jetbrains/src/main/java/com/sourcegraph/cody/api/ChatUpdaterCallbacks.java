package com.sourcegraph.cody.api;

import com.intellij.openapi.diagnostic.Logger;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ChatUpdaterCallbacks implements CompletionsCallbacks {
  private static final Logger logger = Logger.getInstance(ChatUpdaterCallbacks.class);
  private final UpdatableChat chat;
  private final String prefix;
  private boolean gotFirstMessage = false;

  public ChatUpdaterCallbacks(@NotNull UpdatableChat chat, @Nullable String prefix) {
    this.chat = chat;
    this.prefix = prefix;
  }

  @Override
  public void onSubscribed() {
    logger.info("Subscribed to completions.");
  }

  @Override
  public void onData(@Nullable String data) {
    if (data == null) {
      return;
    }
    // print date/time and msg
    // logger.info(DateTimeFormatter.ofPattern("yyyy-MM-dd
    // HH:mm:ss.SSS").format(LocalDateTime.now()) + " Data received by callback: " + data);
    if (!gotFirstMessage) {
      chat.addMessageToChat(ChatMessage.createAssistantMessage(reformatBotMessage(data, prefix)));
      gotFirstMessage = true;
    } else {
      chat.updateLastMessage(ChatMessage.createAssistantMessage(reformatBotMessage(data, prefix)));
    }
  }

  @Override
  public void onError(@NotNull Throwable error) {
    String message = error.getMessage();
    chat.respondToErrorFromServer(message != null ? message : "");
    chat.finishMessageProcessing();
    logger.error("Error: " + error);
  }

  @Override
  public void onComplete() {
    logger.info("Streaming completed.");
    chat.finishMessageProcessing();
  }

  @Override
  public void onCancelled() {
    chat.finishMessageProcessing();
  }

  private static @NotNull String reformatBotMessage(@NotNull String text, @Nullable String prefix) {
    String STOP_SEQUENCE_REGEXP = "(H|Hu|Hum|Huma|Human|Human:)$";
    Pattern stopSequencePattern = Pattern.compile(STOP_SEQUENCE_REGEXP);

    String reformattedMessage = (prefix != null ? prefix : "") + text.stripTrailing();

    Matcher stopSequenceMatcher = stopSequencePattern.matcher(reformattedMessage);

    if (stopSequenceMatcher.find()) {
      reformattedMessage = reformattedMessage.substring(0, stopSequenceMatcher.start());
    }

    return reformattedMessage;
  }
}
