package com.sourcegraph.website;

import com.intellij.openapi.actionSystem.ActionUpdateThread;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.application.ApplicationInfo;
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
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.git.CommitViewUriBuilder;
import com.sourcegraph.git.GitUtil;
import com.sourcegraph.git.RevisionContext;
import org.jetbrains.annotations.NotNull;

import java.net.URI;
import java.util.Optional;

/**
 * JetBrains IDE action to open a selected revision in Sourcegraph.
 */
public class OpenRevisionAction extends DumbAwareAction {
    private final Logger logger = Logger.getInstance(this.getClass());

    @Override
    public @NotNull ActionUpdateThread getActionUpdateThread() {
        return super.getActionUpdateThread();
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

    @Override
    public void actionPerformed(@NotNull AnActionEvent event) {
        // This action handles events for both log and history views, so attempt to load from any possible option.
        RevisionContext context = getHistoryRevisionContext(event).or(() -> getLogRevisionContext(event))
            .orElseThrow(() -> new RuntimeException("Unable to determine revision from history or log."));

        Project project = context.getProject();

        if (project.getProjectFilePath() == null) {
            logger.warn("No project file path found (project: " + project.getName() + ")");
            return;
        }

        String productName = ApplicationInfo.getInstance().getVersionName();
        String productVersion = ApplicationInfo.getInstance().getFullVersion();
        String remoteUrl;
        try {
            remoteUrl = GitUtil.getRemoteUrl(context.getRepoRoot(), project);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
        CommitViewUriBuilder builder = new CommitViewUriBuilder();
        URI uri;
        try {
            uri = builder.build(ConfigUtil.getSourcegraphUrl(project), context.getRevisionNumber(), remoteUrl, productName, productVersion);
        } catch (IllegalArgumentException e) {
            logger.warn("Unable to build commit view URI for url " + ConfigUtil.getSourcegraphUrl(project)
                + ", revision " + context.getRevisionNumber() + ", product " + productName + ", version " + productVersion, e);
            return;
        }
        BrowserOpener.openInBrowser(project, uri);
    }

    @Override
    public void update(@NotNull AnActionEvent event) {
        event.getPresentation().setEnabledAndVisible(true);
    }
}
