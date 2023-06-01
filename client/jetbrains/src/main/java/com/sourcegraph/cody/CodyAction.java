package com.sourcegraph.cody;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import org.jetbrains.annotations.NotNull;

public class CodyAction extends DumbAwareAction {
  @Override
  public void actionPerformed(@NotNull AnActionEvent e) {
    Notification notification =
        new Notification("Cody errors", "Cody", "Test", NotificationType.WARNING);
    AnAction dismissAction =
        new DumbAwareAction("Dismiss") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            notification.expire();
          }
        };
    notification.setIcon(Icons.CodyLogo);
    notification.addAction(dismissAction);
    Notifications.Bus.notify(notification);
    notification.notify(e.getProject());
  }
}
