package com.sourcegraph.common;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.ide.CopyPasteManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.Icons;
import java.awt.datatransfer.StringSelection;
import java.net.URI;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class BrowserErrorNotification {
  public static void show(@Nullable Project project, URI uri) {
    Notification notification =
        new Notification(
            "Sourcegraph errors",
            "Sourcegraph",
            "Opening an external browser is not supported. You can still copy the URL to your clipboard and open it manually.",
            NotificationType.WARNING);
    AnAction copyUrlAction =
        new DumbAwareAction("Copy URL") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            CopyPasteManager.getInstance().setContents(new StringSelection(uri.toString()));
            notification.expire();
          }
        };
    AnAction dismissAction =
        new DumbAwareAction("Dismiss") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            notification.expire();
          }
        };
    notification.setIcon(Icons.CodyLogo);
    notification.addAction(copyUrlAction);
    notification.addAction(dismissAction);
    Notifications.Bus.notify(notification);
    notification.notify(project);
  }
}
