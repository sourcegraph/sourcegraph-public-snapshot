package com.sourcegraph.cody.recipes;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.config.AccountType;
import com.sourcegraph.cody.config.CodyAuthenticationManager;
import com.sourcegraph.cody.localapp.LocalAppManager;
import org.jetbrains.annotations.NotNull;

public class RunCodyAppAction extends DumbAwareAction {
  @Override
  public void actionPerformed(@NotNull AnActionEvent e) {
    LocalAppManager.runLocalApp();
  }

  @Override
  public void update(@NotNull AnActionEvent e) {
    Project project = e.getProject();
    if (project == null) {
      hideAction(e);
      return;
    }
    if (LocalAppManager.isPlatformSupported()
        && CodyAuthenticationManager.getInstance().getActiveAccountType(project)
            == AccountType.LOCAL_APP) {
      if (LocalAppManager.isLocalAppInstalled() && !LocalAppManager.isLocalAppRunning()) {
        showAction(e);
        return;
      }
    }
    hideAction(e);
  }

  private static void showAction(@NotNull AnActionEvent e) {
    e.getPresentation().setEnabledAndVisible(true);
  }

  private static void hideAction(@NotNull AnActionEvent e) {
    e.getPresentation().setEnabledAndVisible(false);
  }
}
