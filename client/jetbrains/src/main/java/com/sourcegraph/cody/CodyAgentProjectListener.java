package com.sourcegraph.cody;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.project.ProjectManagerListener;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;

public class CodyAgentProjectListener implements ProjectManagerListener {
  @Override
  public void projectOpened(@NotNull Project project) {
    if (!ConfigUtil.isCodyEnabled()) {
      return;
    }
    this.startAgent(project);
  }

  @Override
  public void projectClosed(@NotNull Project project) {
    this.stopAgent(project);
  }

  public void stopAgent(@NotNull Project project) {
    CodyAgent service = project.getService(CodyAgent.class);
    if (service == null) {
      return;
    }
    service.shutdown();
  }

  public void startAgent(@NotNull Project project) {
    CodyAgent service = project.getService(CodyAgent.class);
    if (service == null) {
      return;
    }
    if (CodyAgent.isConnected(project)) {
      return;
    }
    service.initialize();
  }
}
