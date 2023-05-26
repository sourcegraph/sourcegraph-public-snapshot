package com.sourcegraph.cody.completions;

import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ChatUpdaterCallbacks implements CompletionsCallbacks {
  private final UpdatableChat chat;
  private final String prefix;
  private boolean gotFirstMessage = false;

  public ChatUpdaterCallbacks(@NotNull UpdatableChat chat, @Nullable String prefix) {
    this.chat = chat;
    this.prefix = prefix;
  }

  @Override
  public void onSubscribed() {
    System.out.println("Subscribed to completions.");
  }

  @Override
  public void onData(@Nullable String data) {
    if (data == null) {
      return;
    }
    // print date/time and msg
    // System.out.println(DateTimeFormatter.ofPattern("yyyy-MM-dd
    // HH:mm:ss.SSS").format(LocalDateTime.now()) + " Data received by callback: " + data);
    if (!gotFirstMessage) {
      chat.addMessage(ChatMessage.createAssistantMessage(reformatBotMessage(data, prefix)));
      gotFirstMessage = true;
    } else {
      chat.updateLastMessage(ChatMessage.createAssistantMessage(reformatBotMessage(data, prefix)));
    }
  }

  @Override
  public void onError(@NotNull Throwable error) {
    if (error.getMessage().equals("Connection refused")) {
      chat.addMessage(
          ChatMessage.createAssistantMessage(
              "I'm sorry, I can't connect to the server. Please make sure that the server is running and try again."));
    } else {
      chat.addMessage(
          ChatMessage.createAssistantMessage(
              "I'm sorry, something wet wrong. Please try again. The error message I got was: \""
                  + error.getMessage()
                  + "\"."));
    }
    System.err.println("Error: " + error);
  }

  @Override
  public void onComplete() {
    System.out.println("Streaming completed.");
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
