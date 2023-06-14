package com.sourcegraph.agent.protocol;

import org.jetbrains.annotations.NotNull;

public class StaticEditor {
  @NotNull public final String workspaceRoot;

  public StaticEditor(@NotNull String workspaceRoot) {
    this.workspaceRoot = workspaceRoot;
  }

  @Override
  public String toString() {
    return "StaticEditor{" + "workspaceRoot='" + workspaceRoot + '\'' + '}';
  }
}
