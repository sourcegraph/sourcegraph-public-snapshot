package com.sourcegraph.config;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.options.ShowSettingsUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.startup.StartupActivity;
import com.sourcegraph.Icons;
import org.jetbrains.annotations.NotNull;

public class NotificationActivity implements StartupActivity.DumbAware {
    @Override
    public void runActivity(@NotNull Project project) {
        String url = ConfigUtil.getSourcegraphUrl(project);
        if (ConfigUtil.isUrlNotificationDismissed() || (url.length() != 0 && !url.startsWith("https://sourcegraph.com"))) {
            return;
        }
        // Display notification
        Notification notification = new Notification("Sourcegraph", "Sourcegraph",
            "A custom Sourcegraph URL is not set for this project. You can only access public repos. Do you want to set your custom URL?", NotificationType.INFORMATION);
        AnAction setUrlAction = new DumbAwareAction("Set URL") {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                ShowSettingsUtil.getInstance().showSettingsDialog(project, SettingsConfigurable.class);
            }
        };
        AnAction cancelAction = new DumbAwareAction("Do Not Set") {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                notification.expire();
            }
        };
        AnAction neverShowAgainAction = new DumbAwareAction("Never Show Again") {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                notification.expire();
                ConfigUtil.setUrlNotificationDismissed(true);
            }
        };
        notification.setIcon(Icons.SourcegraphLogo);
        notification.addAction(setUrlAction);
        notification.addAction(cancelAction);
        notification.addAction(neverShowAgainAction);
        Notifications.Bus.notify(notification);
    }
}
