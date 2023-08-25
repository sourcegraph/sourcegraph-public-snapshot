package com.sourcegraph.config;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.keymap.KeymapUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.startup.StartupActivity;
import com.sourcegraph.Icons;
import com.sourcegraph.common.BrowserOpener;
import com.sourcegraph.find.FindService;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.util.Optional;
import javax.swing.*;
import org.jetbrains.annotations.NotNull;

public class NotificationActivity implements StartupActivity.DumbAware {
  private static AnAction dumbAwareAction(String text, Runnable action) {
    return new DumbAwareAction(text) {
      @Override
      public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
        action.run();
      }
    };
  }

  @Override
  public void runActivity(@NotNull Project project) {
    String latestReleaseMilestoneVersion = "2.0.0";
    String lastNotifiedPluginVersion = ConfigUtil.getLastUpdateNotificationPluginVersion();
    if (lastNotifiedPluginVersion == null
        || lastNotifiedPluginVersion.compareTo(latestReleaseMilestoneVersion) < 0) {
      notifyAboutUpdate(project);
    } else {
      String url = ConfigUtil.getEnterpriseUrl(project);
      if (!ConfigUtil.isUrlNotificationDismissed()
          && (url.length() == 0 || url.startsWith(ConfigUtil.DOTCOM_URL))) {
        notifyAboutSourcegraphUrl();
      }
    }
  }

  private void notifyAboutSourcegraphUrl() {
    // Display notification
    Notification notification =
        new Notification(
            "Sourcegraph: server access",
            "Sourcegraph",
            "A custom Sourcegraph URL is not set for this project. You can only access public repos. Do you want to set your custom URL?",
            NotificationType.INFORMATION);
    AnAction setUrlAction = new OpenPluginSettingsAction("Set URL");
    AnAction cancelAction =
        new DumbAwareAction("Do Not Set") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            notification.expire();
          }
        };
    AnAction neverShowAgainAction =
        new DumbAwareAction("Never Show Again") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            notification.expire();
            ConfigUtil.setUrlNotificationDismissed(true);
          }
        };
    notification.setIcon(Icons.CodyLogo);
    notification.addAction(setUrlAction);
    notification.addAction(cancelAction);
    notification.addAction(neverShowAgainAction);
    Notifications.Bus.notify(notification);
  }

  private void notifyAboutUpdate(@NotNull Project project) {
    // Display notification
    KeyboardShortcut altSShortcut =
        new KeyboardShortcut(KeyStroke.getKeyStroke(KeyEvent.VK_S, InputEvent.ALT_DOWN_MASK), null);
    String altSShortcutText = KeymapUtil.getShortcutText(altSShortcut);
    Notification notification =
        new Notification(
            "Sourcegraph Cody + Code Search plugin updates",
            "Sourcegraph Cody + Code Search",
            "Access the new plugin and try out code search with the shortcut "
                + altSShortcutText
                + "! Learn more about the plugin’s functionality in our blog post.",
            NotificationType.INFORMATION);
    AnAction openAction =
        new DumbAwareAction("Open Sourcegraph (" + altSShortcutText + ")") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            project.getService(FindService.class).showPopup();
            notification.expire();
          }
        };
    AnAction learnMoreAction =
        new DumbAwareAction("Learn More", "Opens browser to describe the latest changes", null) {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            String whatsNewUrl = "https://about.sourcegraph.com/blog/jetbrains-plugin";

            BrowserOpener.openInBrowser(project, whatsNewUrl);
          }
        };
    notification.setIcon(Icons.CodyLogo);
    notification.addAction(openAction);
    notification.addAction(learnMoreAction);
    Notifications.Bus.notify(notification);

    ConfigUtil.setLastUpdateNotificationPluginVersionToCurrent();
  }

  public static void notifyAboutSourcegraphAccessToken(Optional<String> instanceUrl) {
    String content =
        instanceUrl
                .map(url -> "A Sourcegraph Access Token for " + url + " has not been set.")
                .orElse("A Sourcegraph Access Token has not been set.")
            + " It is not possible to use Sourcegraph API and Cody without it.";
    Notification notification =
        new Notification(
            "Sourcegraph: server access", "Sourcegraph", content, NotificationType.INFORMATION);
    notification.setIcon(Icons.CodyLogo);
    notification.addAction(new OpenPluginSettingsAction("Set Access Token"));
    notification.addAction(dumbAwareAction("Do Not Set", notification::expire));
    notification.addAction(
        dumbAwareAction(
            "Never Show Again",
            () -> {
              notification.expire();
              ConfigUtil.setAccessTokenNotificationDismissed(true);
            }));
    Notifications.Bus.notify(notification);
  }
}
