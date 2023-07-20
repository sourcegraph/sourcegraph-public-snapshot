package com.sourcegraph.cody.recipes;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
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
    if (!ConfigUtil.isCodyEnabled()) {
      disableAndHide(e);
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
    enableAndShow(e);
  }

  private static void disableAndHide(@NotNull AnActionEvent e) {
    e.getPresentation().setEnabledAndVisible(false);
  }

  private static void enableAndShow(@NotNull AnActionEvent e) {
    e.getPresentation().setEnabledAndVisible(true);
  }

  private static void disableAction(@NotNull AnActionEvent e) {
    e.getPresentation().setEnabled(false);
  }
}
