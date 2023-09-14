package com.sourcegraph.cody.chat;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.CodyToolWindowContent;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;

public class RefreshCodyAppDetection extends DumbAwareAction {

  @Override
  public void actionPerformed(@NotNull AnActionEvent e) {
    Project project = e.getProject();
    if (project == null) {
      return;
    }
    CodyToolWindowContent codyToolWindowContent = CodyToolWindowContent.getInstance(project);
    if (codyToolWindowContent != null) {
      codyToolWindowContent.refreshPanelsVisibility();
    }
  }

  /**
   * This action is being updated in the background by the intellij action system, and we're using
   * it to show one of the selected panels in the Cody
   */
  @Override
  public void update(@NotNull AnActionEvent e) {
    e.getPresentation().setVisible(false);
    if (!ConfigUtil.isCodyEnabled()) {
      e.getPresentation().setEnabled(false);
      return;
    } else {
      e.getPresentation().setEnabled(true);
    }
    Project project = e.getProject();
    if (project != null) {
      CodyToolWindowContent codyToolWindowContent = CodyToolWindowContent.getInstance(project);
      if (codyToolWindowContent != null) {
        codyToolWindowContent.refreshPanelsVisibility();
      }
    }
  }
}
