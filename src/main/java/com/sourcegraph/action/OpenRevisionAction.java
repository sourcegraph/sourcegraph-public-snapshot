package com.sourcegraph.action;

import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.application.ApplicationInfo;
import com.intellij.openapi.project.DumbAware;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vcs.VcsDataKeys;
import com.intellij.openapi.vcs.history.VcsFileRevision;
import com.intellij.vcs.log.VcsLog;
import com.intellij.vcs.log.VcsLogDataKeys;

import java.awt.Desktop;
import java.io.IOException;
import java.net.URI;
import java.util.Optional;

import com.sourcegraph.project.CommitViewUriBuilder;
import com.sourcegraph.project.RepoInfo;
import com.sourcegraph.project.RevisionContext;
import com.sourcegraph.util.SourcegraphUtil;

import org.jetbrains.annotations.NotNull;
import com.intellij.openapi.diagnostic.Logger;

/**
 * Jetbrains IDE action to open a selected revision in Sourcegraph.
 */
public class OpenRevisionAction extends AnAction implements DumbAware {
  private final Logger logger = Logger.getInstance(this.getClass());

  private Optional<RevisionContext> getHistoryRevision(AnActionEvent e) {
    VcsFileRevision revision = e.getDataContext().getData(VcsDataKeys.VCS_FILE_REVISION);
    Project project = e.getProject();

    if (project == null) {
      return Optional.empty();
    }
    if (revision == null) {
      return Optional.empty();
    }

    String rev = revision.getRevisionNumber().toString();
    return Optional.of(new RevisionContext(project, rev));
  }

  private Optional<RevisionContext> getLogRevision(AnActionEvent e) {
    VcsLog log = e.getDataContext().getData(VcsLogDataKeys.VCS_LOG);
    Project project = e.getProject();

    if (project == null) {
      return Optional.empty();
    }
    if (log == null || log.getSelectedCommits().isEmpty()) {
      return Optional.empty();
    }


        String rev = log.getSelectedCommits().get(0).getHash().asString();
        return Optional.of(new RevisionContext(project, rev));
    }

    @Override
    public void actionPerformed(@NotNull AnActionEvent e) {
        // This action handles events for both log and history views, so attempt to load from any possible option.
        RevisionContext context = getHistoryRevision(e).or(() -> getLogRevision(e))
                .orElseThrow(() -> new RuntimeException("Unable to determine revision from history or log."));

        try {
            String productName = ApplicationInfo.getInstance().getVersionName();
            String productVersion = ApplicationInfo.getInstance().getFullVersion();
            RepoInfo repoInfo = SourcegraphUtil.repoInfo(context.getProject().getProjectFilePath(), context.getProject());

            CommitViewUriBuilder builder = new CommitViewUriBuilder();
            URI uri = builder.build(SourcegraphUtil.sourcegraphURL(context.getProject()), context.getRevisionNumber(), repoInfo, productName, productVersion);

            // Open the URL in the browser.
            Desktop.getDesktop().browse(uri);
        } catch (IOException err) {
            logger.debug("failed to open browser");
            err.printStackTrace();
        }
    }

    @Override
    public void update(@NotNull AnActionEvent e) {
        e.getPresentation().setEnabledAndVisible(true);
    }
}
