package com.sourcegraph.cody;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.project.ProjectManagerListener;
import com.sourcegraph.cody.agent.CodyAgent;
import org.jetbrains.annotations.NotNull;

public class CodyAgentProjectListener implements ProjectManagerListener {
  @Override
  public void projectOpened(@NotNull Project project) {
    CodyAgent service = project.getService(CodyAgent.class);
    if (service == null) {
      return;
    }
    service.initialize();
  }

  @Override
  public void projectClosed(@NotNull Project project) {
    CodyAgent service = project.getService(CodyAgent.class);
    if (service == null) {
      return;
    }
    service.shutdown();
  }
}
