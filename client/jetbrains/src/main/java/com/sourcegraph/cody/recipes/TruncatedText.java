package com.sourcegraph.cody.recipes;

import org.jetbrains.annotations.NotNull;

public class TruncatedText {

  private final @NotNull String value;

  public TruncatedText(@NotNull String value) {
    this.value = value;
  }

  public @NotNull String getValue() {
    return value;
  }
}
