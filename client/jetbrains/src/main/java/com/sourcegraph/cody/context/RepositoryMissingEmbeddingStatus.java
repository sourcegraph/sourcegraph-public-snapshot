package com.sourcegraph.cody.context;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.Icons;
import com.sourcegraph.cody.config.AccountType;
import com.sourcegraph.cody.config.CodyAuthenticationManager;
import javax.swing.Icon;
import org.jetbrains.annotations.NotNull;

public class RepositoryMissingEmbeddingStatus extends RepoAvailableEmbeddingStatus {
  public RepositoryMissingEmbeddingStatus(String repoName) {
    super(repoName);
  }

  @Override
  public Icon getIcon() {
    return Icons.Repository.Missing;
  }

  @Override
  public @NotNull String getTooltip(@NotNull Project project) {
    AccountType accountType =
        CodyAuthenticationManager.getInstance().getDefaultAccountType(project);
    if (accountType == AccountType.LOCAL_APP) {
      return "Repository is not set up in Cody App";
    } else {
      return "Repository does not exist on Sourcegraph";
    }
  }
}
