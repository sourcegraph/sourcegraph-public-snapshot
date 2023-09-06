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
    startAgent(project);
  }

  @Override
  public void projectClosed(@NotNull Project project) {
    stopAgent(project);
  }

  public static void stopAgent(@NotNull Project project) {
    if (project.isDisposed()) {
      return;
    }
    CodyAgent service = project.getService(CodyAgent.class);
    if (service == null) {
      return;
    }
    service.shutdown();
  }

  public static void startAgent(@NotNull Project project) {
    if (project.isDisposed()) {
      return;
    }
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
