package com.sourcegraph.common;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.Icons;
import org.jetbrains.annotations.NotNull;

public class ErrorNotification {
    public static void show(Project project, String errorMessage) {
        Notification notification = new Notification("Sourcegraph errors", "Sourcegraph", errorMessage, NotificationType.WARNING);
        AnAction dismissAction = new DumbAwareAction("Dismiss") {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                notification.expire();
            }
        };
        notification.setIcon(Icons.SourcegraphLogo);
        notification.addAction(dismissAction);
        Notifications.Bus.notify(notification);
        notification.notify(project);
    }
}
