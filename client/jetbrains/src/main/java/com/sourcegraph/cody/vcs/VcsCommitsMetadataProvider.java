package com.sourcegraph.cody.vcs;

import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.vcs.log.*;
import com.intellij.vcs.log.graph.GraphCommit;
import com.intellij.vcs.log.impl.VcsProjectLog;
import java.util.List;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;

class VcsCommitsMetadataProvider {

  private static final @NotNull Logger logger =
      Logger.getInstance(VcsCommitsMetadataProvider.class);

  static String getVcsCommitsMetadataDescription(
      @NotNull Project project, @NotNull VcsFilter vcsFilter) {
    CommitMetadataConsumer commitMetadataConsumer = new CommitMetadataConsumer();

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
}
