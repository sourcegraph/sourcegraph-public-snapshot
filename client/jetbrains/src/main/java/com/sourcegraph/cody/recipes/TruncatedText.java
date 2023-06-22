package com.sourcegraph.cody.recipes;

import static com.sourcegraph.cody.TruncationUtils.CHARS_PER_TOKEN;

import org.jetbrains.annotations.NotNull;

public class TruncatedText {

  private final @NotNull String value;

  public TruncatedText(@NotNull String value, int maxTokens, boolean truncateStart) {
    this.value =
        truncateStart ? truncateTextStart(value, maxTokens) : truncateText(value, maxTokens);
  }

  public static TruncatedText of(@NotNull String value, int maxTokens) {
    return new TruncatedText(value, maxTokens, false);
  }

  public static TruncatedText ofEndOf(@NotNull String value, int maxTokens) {
    return new TruncatedText(value, maxTokens, true);
  }

  public @NotNull String getValue() {
    return value;
  }

  /** Truncates text to the given number of tokens, keeping the start of the text. */
  public static String truncateText(@NotNull String text, int maxTokens) {
    int maxLength = maxTokens * CHARS_PER_TOKEN;
    return text.length() <= maxLength ? text : text.substring(0, maxLength);
  }

  /** Truncates text to the given number of tokens, keeping the end of the text. */
  public static String truncateTextStart(@NotNull String text, int maxTokens) {
    int maxLength = maxTokens * CHARS_PER_TOKEN;
    return text.length() <= maxLength ? text : text.substring(text.length() - maxLength);
  }
}
