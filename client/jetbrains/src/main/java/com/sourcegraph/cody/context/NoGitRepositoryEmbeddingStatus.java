package com.sourcegraph.cody.context;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.Icons;
import javax.swing.Icon;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class NoGitRepositoryEmbeddingStatus implements EmbeddingStatus {
  @Override
  public @Nullable Icon getIcon() {
    return Icons.Repository.Missing;
  }

  @Override
  public @NotNull String getTooltip(@NotNull Project project) {
    return "No Git repository opened";
  }

  @Override
  public @NotNull String getMainText() {
    return "";
  }
}
