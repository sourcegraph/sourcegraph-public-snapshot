package com.sourcegraph.cody.config;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.startup.StartupActivity;
import com.sourcegraph.cody.Icons;
import org.jetbrains.annotations.NotNull;

public class NotificationActivity implements StartupActivity.DumbAware {
    @Override
    public void runActivity(@NotNull Project project) {
        String url = ConfigUtil.getEnterpriseUrl(project);
        if (url.length() == 0 || url.startsWith(ConfigUtil.DOTCOM_URL)) {
            notifyAboutSourcegraphUrl();
        }
    }

    private void notifyAboutSourcegraphUrl() {
        // Display notification
        Notification notification = new Notification("Sourcegraph access", "Cody",
            "A custom Sourcegraph URL is not set. Cody can only access public repos. Do you want to set your custom URL?", NotificationType.INFORMATION);
        AnAction setUrlAction = new OpenPluginSettingsAction("Set URL");
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
                ConfigUtil.setAreChatPredictionsEnabled(true);
            }
        };
        notification.setIcon(Icons.CodyLogo);
        notification.addAction(setUrlAction);
        notification.addAction(cancelAction);
        notification.addAction(neverShowAgainAction);
        Notifications.Bus.notify(notification);
    }
}
