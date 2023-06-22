package com.sourcegraph.cody;

public class TruncationUtils {
  public static final int CHARS_PER_TOKEN = 4;
  public static final int MAX_PROMPT_TOKEN_LENGTH = 7000;
  public static final int SOLUTION_TOKEN_LENGTH = 1000;
  public static final int MAX_HUMAN_INPUT_TOKENS = 1000;
  public static final int MAX_RECIPE_INPUT_TOKENS = 2000;
  public static final int MAX_CURRENT_FILE_TOKENS = 1000;
  public static final int MAX_RECIPE_SURROUNDING_TOKENS = 500;
  public static final int MAX_AVAILABLE_PROMPT_LENGTH =
      MAX_PROMPT_TOKEN_LENGTH - SOLUTION_TOKEN_LENGTH;

  /** The number of code lines to include in the preceding and following texts near the selection */
  public static final int SURROUNDING_LINES = 50;

  public static String truncateText(String text, int maxTokens) {
    int maxLength = maxTokens * CHARS_PER_TOKEN;
    return text.length() <= maxLength ? text : text.substring(0, maxLength);
  }
}
