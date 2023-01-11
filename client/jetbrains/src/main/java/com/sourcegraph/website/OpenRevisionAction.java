package com.sourcegraph.website;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.application.ApplicationInfo;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vcs.VcsDataKeys;
import com.intellij.openapi.vcs.history.VcsFileRevision;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.vcs.log.VcsLog;
import com.intellij.vcs.log.VcsLogDataKeys;
import com.intellij.vcsUtil.VcsUtil;
import com.sourcegraph.common.BrowserOpener;
import com.sourcegraph.common.ErrorNotification;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.vcs.RepoUtil;
import com.sourcegraph.vcs.RevisionContext;
import com.sourcegraph.vcs.VCSType;
import org.jetbrains.annotations.NotNull;

import java.util.Optional;

/**
 * JetBrains IDE action to open a selected revision in Sourcegraph.
 */
@SuppressWarnings("MissingActionUpdateThread")
public class OpenRevisionAction extends DumbAwareAction {
    private final Logger logger = Logger.getInstance(this.getClass());

    @Override
    public void actionPerformed(@NotNull AnActionEvent event) {
        Project project = event.getProject();
        if (project == null) {
            return;
        }

        // This action handles events for both log and history views, so attempt to load from any possible option.
        RevisionContext context = getHistoryRevisionContext(event).or(() -> getLogRevisionContext(event)).orElse(null);

        if (context == null) {
            VirtualFile file = event.getDataContext().getData(VcsDataKeys.VCS_VIRTUAL_FILE);
            if (file != null) {
                // This cannot run on EDT (Event Dispatch Thread) because it may block for a long time.
                ApplicationManager.getApplication().executeOnPooledThread(
                    () -> {
                        if (RepoUtil.getVcsType(project, file) == VCSType.PERFORCE) {
                            // Perforce doesn't have a history view, so we'll just open the file in Sourcegraph.
                            ErrorNotification.show(project, "This feature is not yet supported for Perforce. If you want to see Perforce support sooner than later, please raise this at support@sourcegraph.com.");
                        } else {
                            ErrorNotification.show(project, "Could not find revision to open.");
                        }
                    });
            } else {
                ErrorNotification.show(project, "Could not find revision to open.");
            }
            return;
        }

        if (project.getProjectFilePath() == null) {
            ErrorNotification.show(project, "No project file path found (project: " + project.getName() + ")");
            return;
        }

        String productName = ApplicationInfo.getInstance().getVersionName();
        String productVersion = ApplicationInfo.getInstance().getFullVersion();

        // This cannot run on EDT (Event Dispatch Thread) because it may block for a long time.
        ApplicationManager.getApplication().executeOnPooledThread(
            () -> {
                String remoteUrl;
                try {
                    remoteUrl = RepoUtil.getRemoteRepoUrl(project, context.getRepoRoot());
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }

                String url;
                try {
                    url = URLBuilder.buildCommitUrl(ConfigUtil.getSourcegraphUrl(project), context.getRevisionNumber(), remoteUrl, productName, productVersion);
                } catch (IllegalArgumentException e) {
                    logger.warn("Unable to build commit view URI for url " + ConfigUtil.getSourcegraphUrl(project)
                        + ", revision " + context.getRevisionNumber() + ", product " + productName + ", version " + productVersion, e);
                    return;
                }
                BrowserOpener.openInBrowser(project, url);
            }
        );
    }

    @Override
    public void update(@NotNull AnActionEvent event) {
        event.getPresentation().setEnabledAndVisible(true);
    }

    @NotNull
    private Optional<RevisionContext> getHistoryRevisionContext(@NotNull AnActionEvent event) {
        Project project = event.getProject();
        VcsFileRevision revisionObject = event.getDataContext().getData(VcsDataKeys.VCS_FILE_REVISION);
        VirtualFile file = event.getDataContext().getData(VcsDataKeys.VCS_VIRTUAL_FILE);

        if (project == null || revisionObject == null || file == null) {
            return Optional.empty();
        }

        String revision = revisionObject.getRevisionNumber().toString();
        VirtualFile root = VcsUtil.getVcsRootFor(project, file);
        if (root == null) {
            return Optional.empty();
        }
        return Optional.of(new RevisionContext(project, revision, root));
    }

    @NotNull
    private Optional<RevisionContext> getLogRevisionContext(@NotNull AnActionEvent event) {
        VcsLog log = event.getDataContext().getData(VcsLogDataKeys.VCS_LOG);
        Project project = event.getProject();

        if (project == null) {
            return Optional.empty();
        }
        if (log == null || log.getSelectedCommits().isEmpty()) {
            return Optional.empty();
        }

        String revision = log.getSelectedCommits().get(0).getHash().asString();
        VirtualFile root = log.getSelectedCommits().get(0).getRoot();
        return Optional.of(new RevisionContext(project, revision, root));
    }
}
