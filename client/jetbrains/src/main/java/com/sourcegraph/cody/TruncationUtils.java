package com.sourcegraph.cody;

import org.jetbrains.annotations.NotNull;

public class TruncationUtils {
  public static final int CHARS_PER_TOKEN = 4;
  public static final int SOLUTION_TOKEN_LENGTH = 1000;
  public static final int MAX_HUMAN_INPUT_TOKENS = 1000;
  public static final int MAX_RECIPE_INPUT_TOKENS = 2000;
  public static final int MAX_CURRENT_FILE_TOKENS = 1000;
  public static final int MAX_RECIPE_SURROUNDING_TOKENS = 500;

  /** The number of code lines to include in the preceding and following texts near the selection */
  public static final int SURROUNDING_LINES = 50;

  /** Truncates text to the given number of tokens, keeping the start of the text. */
  public static String truncateText(@NotNull String text, int maxTokens) {
    int maxLength = maxTokens * CHARS_PER_TOKEN;
    return text.length() <= maxLength ? text : text.substring(0, maxLength);
  }
}
