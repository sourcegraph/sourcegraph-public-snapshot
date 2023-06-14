package com.sourcegraph.cody.recipes;

import org.jetbrains.annotations.NotNull;

public class Language {

  private final @NotNull String value;

  public Language(@NotNull String value) {
    this.value = value;
  }

  public @NotNull String getValue() {
    return value;
  }
}
