package com.sourcegraph.cody.context;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.Icons;
import javax.swing.Icon;
import org.jetbrains.annotations.NotNull;

public class RepositoryIndexedEmbeddingStatus extends RepoAvailableEmbeddingStatus {
  public RepositoryIndexedEmbeddingStatus(String repoName) {
    super(repoName);
  }

  @Override
  public Icon getIcon() {
    return Icons.Repository.Indexed;
  }

  @Override
  public @NotNull String getTooltip(@NotNull Project project) {
    return "Repository is indexed";
  }
}
