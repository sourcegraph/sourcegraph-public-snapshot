package com.sourcegraph.cody;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.project.Project;
import com.intellij.ui.TextFieldWithAutoCompletion;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.vcs.RepoUtil;
import git4idea.repo.GitRepository;
import java.util.Collection;
import javax.security.auth.Destroyable;
import org.jetbrains.annotations.NotNull;

public class CodebaseSelector extends TextFieldWithAutoCompletion<String> implements Destroyable {
  public CodebaseSelector(@NotNull Project project) {
    super(project, new StringsCompletionProvider(null, null), true, "");
    refreshRepos();
    // Note: This only works with Git, not with Perforce.
    MessageBusConnection connection = project.getMessageBus().connect();
    connection.subscribe(GitRepository.GIT_REPO_CHANGE, unused -> refreshRepos());
  }

  private void refreshRepos() {
    ApplicationManager.getApplication().executeOnPooledThread(() -> {
      Collection<String> allRepoNames = RepoUtil.getAllRepoNames(this.getProject());
      ApplicationManager.getApplication().invokeLater(() -> this.setVariants(allRepoNames));
    });
  }
}
