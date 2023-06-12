package com.sourcegraph.cody.recipes;

import org.jetbrains.annotations.NotNull;

public class SelectedText {

  private final @NotNull String value;

  public SelectedText(@NotNull String value) {
    this.value = value;
  }

  public @NotNull String getValue() {
    return value;
  }
}
