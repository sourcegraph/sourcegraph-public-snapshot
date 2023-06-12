package com.sourcegraph.cody.recipes;

import org.jetbrains.annotations.NotNull;

public class PromptContext {
  private final @NotNull String prompt;
  private final @NotNull String displayText;
  private final @NotNull String responsePrefix;

  public PromptContext(
      @NotNull String prompt, @NotNull String displayText, @NotNull String responsePrefix) {
    this.prompt = prompt;
    this.displayText = displayText;
    this.responsePrefix = responsePrefix;
  }

  public PromptContext(@NotNull String prompt, @NotNull String displayText) {
    this(prompt, displayText, "");
  }

  public @NotNull String getPrompt() {
    return prompt;
  }

  public @NotNull String getDisplayText() {
    return displayText;
  }

  public @NotNull String getResponsePrefix() {
    return responsePrefix;
  }
}
