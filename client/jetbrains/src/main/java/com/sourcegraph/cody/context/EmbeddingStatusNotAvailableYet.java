package com.sourcegraph.cody.context;

import com.intellij.openapi.project.Project;
import javax.swing.Icon;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class EmbeddingStatusNotAvailableYet implements EmbeddingStatus {
  @Override
  public @Nullable Icon getIcon() {
    return null;
  }

  @Override
  public @NotNull String getTooltip(@NotNull Project project) {
    return "";
  }

  @Override
  public @NotNull String getMainText() {
    return "";
  }
}
