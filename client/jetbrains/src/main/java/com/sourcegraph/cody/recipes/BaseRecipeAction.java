package com.sourcegraph.cody.recipes;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.UpdatableChatHolderService;
import com.sourcegraph.cody.localapp.LocalAppManager;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.SettingsComponent;
import org.jetbrains.annotations.NotNull;

public abstract class BaseRecipeAction extends DumbAwareAction {

  @Override
  public void update(@NotNull AnActionEvent e) {
    Project project = e.getProject();
    if (project == null) {
      disableAction(e);
      return;
    }
    UpdatableChatHolderService updatableChatHolderService =
        project.getService(UpdatableChatHolderService.class);
    UpdatableChat updatableChat = updatableChatHolderService.getUpdatableChat();
    if (updatableChat == null) {
      disableAction(e);
      return;
    }
    if (LocalAppManager.isPlatformSupported()
        && ConfigUtil.getInstanceType(project) == SettingsComponent.InstanceType.LOCAL_APP) {
      if (!LocalAppManager.isLocalAppInstalled()) {
        disableAction(e);
        return;
      } else if (!LocalAppManager.isLocalAppRunning()) {
        disableAction(e);
        return;
      }
    }
    enableAction(e);
  }

  private static void enableAction(@NotNull AnActionEvent e) {
    e.getPresentation().setEnabled(true);
  }

  private static void disableAction(@NotNull AnActionEvent e) {
    e.getPresentation().setEnabled(false);
  }
}
