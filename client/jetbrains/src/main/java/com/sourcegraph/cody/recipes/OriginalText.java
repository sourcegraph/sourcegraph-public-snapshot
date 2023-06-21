package com.sourcegraph.cody.recipes;

import org.jetbrains.annotations.NotNull;

public class OriginalText {

  private final @NotNull String value;

  public OriginalText(@NotNull String value) {
    this.value = value;
  }

  public @NotNull String getValue() {
    return value;
  }
}
