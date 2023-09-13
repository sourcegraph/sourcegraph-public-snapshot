package com.sourcegraph.cody.agent;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.cody.CodyToolWindowContent;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.vcs.RepoUtil;
import java.util.Objects;
import org.jetbrains.annotations.Nullable;

public class CodyAgentCodebase {
  private final CodyAgentServer underlying;
  @Nullable public String currentCodebase = null;

  public CodyAgentCodebase(CodyAgentServer underlying) {
    this.underlying = underlying;
  }

  public void handlePotentialCodebaseChange(Project project, VirtualFile file) {
    ApplicationManager.getApplication()
        .executeOnPooledThread(
            () -> {
              String repositoryName = RepoUtil.findRepositoryName(project, file);
              ApplicationManager.getApplication()
                  .invokeLater(
                      () -> {
                        if (!Objects.equals(this.currentCodebase, repositoryName)) {
                          this.currentCodebase = repositoryName;
                          CodyToolWindowContent.getInstance(project)
                              .embeddingStatusView
                              .updateEmbeddingStatus();
                          underlying.configurationDidChange(
                              ConfigUtil.getAgentConfiguration(project));
                        }
                      });
            });
  }
}
