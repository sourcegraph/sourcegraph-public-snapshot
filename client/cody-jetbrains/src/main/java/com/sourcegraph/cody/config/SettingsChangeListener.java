package com.sourcegraph.cody.config;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.util.messages.MessageBus;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.cody.telemetry.GraphQlLogger;
import java.util.Objects;
import org.jetbrains.annotations.NotNull;

/**
 * Listens to changes in the plugin settings and: - logs install/uninstall events. - notifies the
 * user about a successful connection.
 */
public class SettingsChangeListener implements Disposable {
  private final MessageBusConnection connection;

  public SettingsChangeListener(@NotNull Project project) {
    MessageBus bus = project.getMessageBus();

    connection = bus.connect();
    connection.subscribe(
        PluginSettingChangeActionNotifier.TOPIC,
        new PluginSettingChangeActionNotifier() {
          @Override
          public void beforeAction(@NotNull PluginSettingChangeContext context) {
            if (!Objects.equals(context.oldEnterpriseUrl, context.newEnterpriseUrl)) {
              GraphQlLogger.logUninstallEvent(project);
              ConfigUtil.setInstallEventLogged(false);
            }
          }

          @Override
          public void afterAction(@NotNull PluginSettingChangeContext context) {
            // Log install events
            if (!Objects.equals(context.oldEnterpriseUrl, context.newEnterpriseUrl)) {
              GraphQlLogger.logInstallEvent(project, ConfigUtil::setInstallEventLogged);
            } else if (!Objects.equals(
                    context.oldEnterpriseAccessToken, context.newEnterpriseAccessToken)
                && !ConfigUtil.isInstallEventLogged()) {
              GraphQlLogger.logInstallEvent(project, ConfigUtil::setInstallEventLogged);
            }

            // Notify user about a successful connection
            if (context.newEnterpriseUrl != null) {
              ApiAuthenticator.testConnection(
                  context.newEnterpriseUrl,
                  context.newEnterpriseAccessToken,
                  context.newCustomRequestHeaders,
                  (status) -> {
                    if (ConfigUtil.didAuthenticationFailLastTime()
                        && status == ApiAuthenticator.ConnectionStatus.AUTHENTICATED) {
                      notifyAboutSuccessfulConnection();
                    }
                    ConfigUtil.setAuthenticationFailedLastTime(
                        status != ApiAuthenticator.ConnectionStatus.AUTHENTICATED);
                  });
            }
          }
        });
  }

  private void notifyAboutSuccessfulConnection() {
    Notification notification =
        new Notification(
            "Cody Sourcegraph access",
            "Cody authentication success",
            "Cody successfully connected to your Sourcegraph account.",
            NotificationType.INFORMATION);
    AnAction dismissAction =
        new DumbAwareAction("Dismiss") {
          @Override
          public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
            notification.expire();
          }
        };
    notification.addAction(dismissAction);
    Notifications.Bus.notify(notification);
  }

  @Override
  public void dispose() {
    connection.disconnect();
  }
}
