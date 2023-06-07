package com.sourcegraph.cody.chat;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.components.ServiceManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.UpdatableChatHolderService;
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
    UpdatableChatHolderService updatableChatHolderService =
        ServiceManager.getService(project, UpdatableChatHolderService.class);
    UpdatableChat updatableChat = updatableChatHolderService.getUpdatableChat();
    updatableChat.resetConversation();
  }

  private static void displayUnableToResetConversationError() {
    ErrorNotification.show(
        null,
        "Unable to reset the current conversation with Cody. Please try again or reach out to us at support@sourcegraph.com.");
  }
}
