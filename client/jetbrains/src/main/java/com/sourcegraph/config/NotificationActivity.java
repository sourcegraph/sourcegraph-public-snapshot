package com.sourcegraph.config;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.externalSystem.service.execution.NotSupportedException;
import com.intellij.openapi.keymap.KeymapUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.startup.StartupActivity;
import com.sourcegraph.Icons;
import com.sourcegraph.find.FindService;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;

public class NotificationActivity implements StartupActivity.DumbAware {
    @Override
    public void runActivity(@NotNull Project project) {
        String latestReleaseMilestoneVersion = "2.0.0";
        String lastNotifiedPluginVersion = ConfigUtil.getLastUpdateNotificationPluginVersion();
        if (lastNotifiedPluginVersion == null || lastNotifiedPluginVersion.compareTo(latestReleaseMilestoneVersion) < 0) {
            notifyAboutUpdate(project);
        } else {
            String url = ConfigUtil.getSourcegraphUrl(project);
            if (!ConfigUtil.isUrlNotificationDismissed() && (url.length() == 0 || url.startsWith("https://sourcegraph.com"))) {
                notifyAboutSourcegraphUrl();
            }
        }
    }

    private void notifyAboutSourcegraphUrl() {
        // Display notification
        Notification notification = new Notification("Sourcegraph access", "Sourcegraph",
            "A custom Sourcegraph URL is not set for this project. You can only access public repos. Do you want to set your custom URL?", NotificationType.INFORMATION);
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
                ConfigUtil.setUrlNotificationDismissed(true);
            }
        };
        notification.setIcon(Icons.SourcegraphLogo);
        notification.addAction(setUrlAction);
        notification.addAction(cancelAction);
        notification.addAction(neverShowAgainAction);
        Notifications.Bus.notify(notification);
    }

    private void notifyAboutUpdate(@NotNull Project project) {
        // Display notification
        KeyboardShortcut altSShortcut = new KeyboardShortcut(KeyStroke.getKeyStroke(KeyEvent.VK_S, InputEvent.ALT_DOWN_MASK), null);
        String altEnterShortcutText = KeymapUtil.getShortcutText(altSShortcut);
        Notification notification = new Notification("Sourcegraph plugin updates", "Sourcegraph",
            "Access the new plugin and try out code search with the shortcut " + altEnterShortcutText + "! Learn more about the plugin’s functionality in our blog post.", NotificationType.INFORMATION);
        AnAction setUrlAction = new DumbAwareAction("Open Sourcegraph (" + altEnterShortcutText + ")") {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                FindService service = project.getService(FindService.class);
                service.showPopup();
                notification.expire();
            }
        };
        AnAction learnMoreAction = new DumbAwareAction("Learn More", "Opens browser to describe the latest changes", null) {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                String whatsNewUrl = "https://plugins.jetbrains.com/plugin/9682-sourcegraph#:~:text=What%E2%80%99s%20New";

                if (Desktop.isDesktopSupported() && Desktop.getDesktop().isSupported(Desktop.Action.BROWSE)) {
                    try {
                        Desktop.getDesktop().browse(new URI(whatsNewUrl));
                    } catch (IOException | URISyntaxException e) {
                        throw new NotSupportedException("Can't open link. Wrong URL.");
                    }
                } else {
                    throw new NotSupportedException("Can't open link. Desktop is not supported.");
                }


            }
        };
        notification.setIcon(Icons.SourcegraphLogo);
        notification.addAction(setUrlAction);
        notification.addAction(learnMoreAction);
        Notifications.Bus.notify(notification);

        ConfigUtil.setLastUpdateNotificationPluginVersionToCurrent();
    }
}
