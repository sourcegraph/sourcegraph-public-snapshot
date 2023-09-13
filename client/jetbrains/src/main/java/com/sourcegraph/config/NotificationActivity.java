package com.sourcegraph.config;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.startup.StartupActivity;
import com.sourcegraph.Icons;
import com.sourcegraph.cody.config.AccountType;
import com.sourcegraph.cody.config.CodyApplicationSettings;
import com.sourcegraph.cody.config.CodyAuthenticationManager;
import javax.swing.*;
import org.jetbrains.annotations.NotNull;

public class NotificationActivity implements StartupActivity.DumbAware {

  @Override
  public void runActivity(@NotNull Project project) {
    AccountType defaultAccountType =
        CodyAuthenticationManager.getInstance().getDefaultAccountType(project);
    if (!ConfigUtil.isDefaultDotcomAccountNotificationDismissed()
        && (defaultAccountType == AccountType.DOTCOM)) {
      notifyAboutDefaultDotcomAccount();
    }
  }

  private void notifyAboutDefaultDotcomAccount() {
    // Display notification
    Notification notification =
        new Notification(
            "Sourcegraph: server access",
            "Sourcegraph",
            "An enterprise Sourcegraph account is not set for this project. You can only access public repos. Do you want to set an enterprise account as the default one?",
            NotificationType.INFORMATION);
    AnAction setEnterpriseAccountAction = new OpenPluginSettingsAction("Set Enterprise Account");
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
            CodyApplicationSettings.getInstance()
                .setDefaultDotcomAccountNotificationDismissed(true);
          }
        };
    notification.setIcon(Icons.CodyLogo);
    notification.addAction(setEnterpriseAccountAction);
    notification.addAction(cancelAction);
    notification.addAction(neverShowAgainAction);
    Notifications.Bus.notify(notification);
  }
}
