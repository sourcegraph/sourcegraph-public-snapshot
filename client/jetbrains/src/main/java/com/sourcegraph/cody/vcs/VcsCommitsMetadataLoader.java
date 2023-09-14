package com.sourcegraph.cody.vcs;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.application.ex.ApplicationUtil;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.progress.ProcessCanceledException;
import com.intellij.openapi.progress.ProgressIndicator;
import com.intellij.openapi.progress.ProgressManager;
import com.intellij.openapi.progress.util.ProgressIndicatorBase;
import com.intellij.openapi.project.Project;
import com.sourcegraph.common.ErrorNotification;
import java.util.Optional;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;
import org.jetbrains.annotations.NotNull;

public class VcsCommitsMetadataLoader {
  private static final @NotNull Logger logger = Logger.getInstance(VcsCommitsMetadataLoader.class);
  private final @NotNull Project project;

  public VcsCommitsMetadataLoader(@NotNull Project project) {
    this.project = project;
  }

  public Optional<String> loadCommitsMetadataDescription(@NotNull VcsFilter vcsFilter) {
    Future<String> loadVcsCommitsMetadataDescription =
        ApplicationManager.getApplication()
            .executeOnPooledThread(
                () ->
                    VcsCommitsMetadataProvider.getVcsCommitsMetadataDescription(
                        project, vcsFilter));
    try {
      ProgressIndicator progressIndicator = ProgressManager.getInstance().getProgressIndicator();
      if (progressIndicator == null) {
        progressIndicator = new ProgressIndicatorBase();
      }
      String commitsMetadataDescription =
          ApplicationUtil.runWithCheckCanceled(
              loadVcsCommitsMetadataDescription, progressIndicator);
      return Optional.of(commitsMetadataDescription);
    } catch (ProcessCanceledException e) {
      // process cancelled by the user
      return Optional.empty();
    } catch (ExecutionException e) {
      logger.warn(e.getMessage());
      ErrorNotification.show(
          project,
          "Unable to load history from version control system. Please try again or reach out to us at support@sourcegraph.com.");
      return Optional.empty();
    }
  }
}
