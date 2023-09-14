package com.sourcegraph.cody.vcs;

import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vcs.roots.VcsRootProblemNotifier;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.vcs.log.*;
import com.intellij.vcs.log.graph.GraphCommit;
import com.intellij.vcs.log.impl.VcsProjectLog;
import git4idea.GitUtil;
import git4idea.repo.GitRepository;
import java.util.Collection;
import java.util.List;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;

class VcsCommitsMetadataProvider {

  private static final @NotNull Logger logger =
      Logger.getInstance(VcsCommitsMetadataProvider.class);

  static @NotNull String getVcsCommitsMetadataDescription(
      @NotNull Project project, @NotNull VcsFilter vcsFilter) {
    CommitMetadataConsumer commitMetadataConsumer = new CommitMetadataConsumer();

    VcsRootProblemNotifier.createInstance(project).rescanAndNotifyIfNeeded();

    if (!isGitRepositoryAvailable(project)) {
      return commitMetadataConsumer.getCommitsContent();
    }

    VcsProjectLog.getLogProviders(project)
        .forEach(
            (virtualFileRoot, vcsLogProvider) -> {
              try {
                List<TimedVcsCommit> commits =
                    vcsLogProvider.getCommitsMatchingFilter(
                        virtualFileRoot, vcsFilter.getFilterCollection(), vcsFilter.getCount());
                List<String> commitHashes =
                    commits.stream()
                        .map(GraphCommit::getId)
                        .map(Hash::asString)
                        .collect(Collectors.toList());
                vcsLogProvider.readMetadata(virtualFileRoot, commitHashes, commitMetadataConsumer);
              } catch (Exception ex) {
                logger.warn(ex);
              }
            });
    return commitMetadataConsumer.getCommitsContent();
  }

  public static boolean isGitRepositoryAvailable(@NotNull Project project) {
    @NotNull Collection<GitRepository> repositories = GitUtil.getRepositories(project);
    for (GitRepository repository : repositories) {
      VirtualFile gitDir = repository.getRoot().findChild(".git");
      if (gitDir != null && gitDir.exists()) {
        return true;
      }
    }
    return false;
  }
}
