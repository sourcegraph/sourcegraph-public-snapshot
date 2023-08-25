package com.sourcegraph.cody;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.editor.event.DocumentEvent;
import com.intellij.openapi.editor.event.DocumentListener;
import com.intellij.openapi.project.Project;
import com.intellij.ui.TextFieldWithAutoCompletion;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.config.CodyProjectService;
import com.sourcegraph.vcs.RepoUtil;
import git4idea.repo.GitRepository;
import java.util.Collection;
import javax.security.auth.Destroyable;
import org.jetbrains.annotations.NotNull;

public class CodebaseSelector extends TextFieldWithAutoCompletion<String> implements Destroyable {
  public CodebaseSelector(@NotNull Project project) {
    super(
        project,
        new StringsCompletionProvider(null, null),
        true,
        CodyProjectService.getInstance(project).getCodyCodebase());
    refreshRepos();

    this.setPlaceholder("Codebase is inferred from open file");

    // Set tooltip to something helpful
    this.setToolTipText(
        "Set the codebase to use, in the format of \"github.com/sourcegraph/cody\". If not set, the codebase is inferred from the currently open file.");

    // Note: This only works with Git, not with Perforce.
    MessageBusConnection connection = project.getMessageBus().connect();
    connection.subscribe(GitRepository.GIT_REPO_CHANGE, unused -> refreshRepos());

    this.addDocumentListener(
        new DocumentListener() {
          @Override
          public void documentChanged(@NotNull DocumentEvent event) {
            CodyProjectService.getInstance(project).codyCodebase = event.getDocument().getText();
          }
        });
  }

  private void refreshRepos() {
    ApplicationManager.getApplication()
        .executeOnPooledThread(
            () -> {
              Collection<String> allRepoNames = RepoUtil.getAllRepoNames(this.getProject());
              ApplicationManager.getApplication().invokeLater(() -> this.setVariants(allRepoNames));
            });
  }
}
