package com.sourcegraph.cody;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.project.ProjectManagerListener;
import com.sourcegraph.cody.agent.CodyAgentManager;
import org.jetbrains.annotations.NotNull;

/**
 * Should get migrated to ProjectCloseListener when compatibility version gets bumped (not yet
 * available in 221).
 *
 * <p>Shouldn't override .projectOpened(), it's currently being done in PostStartupActivity.
 *
 * <p>For more context see the <a
 * href="https://plugins.jetbrains.com/docs/intellij/plugin-components.html">Plugin Components
 * migration doc</a>.
 */
public class CodyAgentProjectListener implements ProjectManagerListener {
  @Override
  public void projectClosed(@NotNull Project project) {
    CodyAgentManager.stopAgent(project);
  }
}
