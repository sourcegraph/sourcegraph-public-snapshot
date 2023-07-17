package com.sourcegraph.cody.chat;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.UpdatableChatHolderService;
import org.jetbrains.annotations.NotNull;

public class RefreshCodyAppDetection extends DumbAwareAction {

  @Override
  public void actionPerformed(@NotNull AnActionEvent e) {
    Project project = e.getProject();
    if (project == null) {
      return;
    }
    UpdatableChatHolderService updatableChatHolderService =
        project.getService(UpdatableChatHolderService.class);
    UpdatableChat updatableChat = updatableChatHolderService.getUpdatableChat();
    if (updatableChat != null) {
      updatableChat.refreshPanelsVisibility();
    }
  }

  /**
   * This action is being updated in the background by the intellij action system, and we're using
   * it to show one of the selected panels in the Cody
   */
  @Override
  public void update(@NotNull AnActionEvent e) {
    e.getPresentation().setVisible(false);
    Project project = e.getProject();
    if (project != null) {
      UpdatableChatHolderService updatableChatHolderService =
          project.getService(UpdatableChatHolderService.class);
      UpdatableChat updatableChat = updatableChatHolderService.getUpdatableChat();
      if (updatableChat != null) {
        updatableChat.refreshPanelsVisibility();
      }
    }
  }
}
