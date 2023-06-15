package com.sourcegraph.cody.vcs;

import com.intellij.util.Consumer;
import com.intellij.vcs.log.VcsCommitMetadata;
import org.jetbrains.annotations.NotNull;

public class CommitMetadataConsumer implements Consumer<VcsCommitMetadata> {

  private final @NotNull StringBuilder commitsContent = new StringBuilder();

  @Override
  public void consume(VcsCommitMetadata vcsCommitMetadata) {
    commitsContent
        .append("Commit author: ")
        .append(vcsCommitMetadata.getAuthor().getName())
        .append("\n")
        .append("Commit message: ")
        .append(vcsCommitMetadata.getFullMessage())
        .append("\n\n");
  }

  public String getCommitsContent() {
    return commitsContent.toString();
  }
}
