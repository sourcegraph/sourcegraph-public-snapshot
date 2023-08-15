package com.sourcegraph.cody.chat;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.CodyToolWindowContent;
import com.sourcegraph.common.ErrorNotification;
import org.jetbrains.annotations.NotNull;

public class ResetCurrentConversationAction extends DumbAwareAction {

  @Override
  public void actionPerformed(@NotNull AnActionEvent e) {
    Project project = e.getProject();
    if (project == null) {
      displayUnableToResetConversationError();
      return;
    }
    CodyToolWindowContent codyToolWindowContent = CodyToolWindowContent.getInstance(project);
    if (codyToolWindowContent != null) {
      codyToolWindowContent.resetConversation();
    }
  }

  @Override
  public void update(@NotNull AnActionEvent e) {
    Project project = e.getProject();
    if (project != null) {
      CodyToolWindowContent codyToolWindowContent = CodyToolWindowContent.getInstance(project);
      if (codyToolWindowContent != null) {
        e.getPresentation().setVisible(codyToolWindowContent.isChatVisible());
      }
    }
  }

  private static void displayUnableToResetConversationError() {
    ErrorNotification.show(
        null,
        "Unable to reset the current conversation with Cody. Please try again or reach out to us at support@sourcegraph.com.");
  }
}
