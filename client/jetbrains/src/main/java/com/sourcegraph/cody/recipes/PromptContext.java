package com.sourcegraph.cody.recipes;

import org.jetbrains.annotations.NotNull;

public class PromptContext {
  private final @NotNull String prompt;
  private final @NotNull String displayText;

  public PromptContext(@NotNull String prompt, @NotNull String displayText) {
    this.prompt = prompt;
    this.displayText = displayText;
  }

  public @NotNull String getPrompt() {
    return prompt;
  }

  public @NotNull String getDisplayText() {
    return displayText;
  }
}
