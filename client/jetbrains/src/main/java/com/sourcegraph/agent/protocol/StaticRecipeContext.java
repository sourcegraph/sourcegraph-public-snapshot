package com.sourcegraph.agent.protocol;

import org.jetbrains.annotations.NotNull;

public class StaticRecipeContext {
  @NotNull public final StaticEditor editor;

  public StaticRecipeContext(@NotNull StaticEditor editor) {
    this.editor = editor;
  }

  @Override
  public String toString() {
    return "StaticRecipeContext{" + "editor=" + editor + '}';
  }
}
