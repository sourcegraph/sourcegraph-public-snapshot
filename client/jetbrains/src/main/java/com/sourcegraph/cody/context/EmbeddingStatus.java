package com.sourcegraph.cody.context;

import com.intellij.openapi.project.Project;
import javax.swing.Icon;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public interface EmbeddingStatus {
  @Nullable
  Icon getIcon();

  @NotNull
  String getTooltip(@NotNull Project project);

  @NotNull
  String getMainText();
}
