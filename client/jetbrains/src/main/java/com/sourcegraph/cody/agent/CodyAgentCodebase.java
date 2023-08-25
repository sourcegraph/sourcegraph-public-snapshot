package com.sourcegraph.cody.agent;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.vcs.RepoUtil;
import java.util.Objects;

public class CodyAgentCodebase {
  private final CodyAgentServer underlying;
  private String currentCodebase = null;

  public CodyAgentCodebase(CodyAgentServer underlying) {
    this.underlying = underlying;
  }

  public void handlePotentialCodebaseChange(Project project, VirtualFile file) {
    ApplicationManager.getApplication()
        .executeOnPooledThread(
            () -> {
              String autodetectedRepositoryName = RepoUtil.findRepositoryName(project, file);
              ApplicationManager.getApplication()
                  .invokeLater(
                      () -> {
                        ExtensionConfiguration config = ConfigUtil.getAgentConfiguration(project);
                        if (config.codebase == null || config.codebase.isEmpty()) {
                          config.setCodebase(autodetectedRepositoryName);
                        }
                        if (config.codebase != null && !config.codebase.equals(this.currentCodebase)) {
                          this.currentCodebase = config.codebase;
                          underlying.configurationDidChange(config);
                        }
                      });
            });
  }
}
