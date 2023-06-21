package com.sourcegraph.config;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileEditor.TextEditor;
import com.intellij.openapi.keymap.KeymapUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.project.ProjectManager;
import com.intellij.util.messages.MessageBus;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.cody.completions.CodyCompletionsManager;
import com.sourcegraph.find.browser.JavaToJSBridge;
import com.sourcegraph.telemetry.GraphQlLogger;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.util.Arrays;
import java.util.Objects;
import javax.swing.*;
import org.jetbrains.annotations.NotNull;

/**
 * Listens to changes in the plugin settings and: - notifies JCEF about them. - logs
 * install/uninstall events. - notifies the user about a successful connection.
 */
public class SettingsChangeListener implements Disposable {
  private final MessageBusConnection connection;
  private JavaToJSBridge javaToJSBridge;

  public SettingsChangeListener(@NotNull Project project) {
    MessageBus bus = project.getMessageBus();

    connection = bus.connect();
    connection.subscribe(
        PluginSettingChangeActionNotifier.TOPIC,
        new PluginSettingChangeActionNotifier() {
          @Override
          public void beforeAction(@NotNull PluginSettingChangeContext context) {
            if (!Objects.equals(context.oldUrl, context.newUrl)) {
              GraphQlLogger.logUninstallEvent(project);
              ConfigUtil.setInstallEventLogged(false);
            }
          }

          @Override
          public void afterAction(@NotNull PluginSettingChangeContext context) {
            // Notify JCEF about the config changes
            if (javaToJSBridge != null) {
              javaToJSBridge.callJS("pluginSettingsChanged", ConfigUtil.getConfigAsJson(project));
            }

            // Log install events
            if (!Objects.equals(context.oldUrl, context.newUrl)) {
              GraphQlLogger.logInstallEvent(project, ConfigUtil::setInstallEventLogged);
            } else if ((!Objects.equals(context.oldDotComAccessToken, context.newDotComAccessToken)
                    || !Objects.equals(
                        context.oldEnterpriseAccessToken, context.newEnterpriseAccessToken))
                && !ConfigUtil.isInstallEventLogged()) {
              GraphQlLogger.logInstallEvent(project, ConfigUtil::setInstallEventLogged);
            }

            // Notify user about a successful connection
            if (context.newUrl != null) {
              final String accessToken =
                  ConfigUtil.getInstanceType(project) == SettingsComponent.InstanceType.DOTCOM
                      ? context.newDotComAccessToken
                      : context.newEnterpriseAccessToken;
              ApiAuthenticator.testConnection(
                  context.newUrl,
                  accessToken,
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

            // clear completions if freshly disabled
            if (context.oldCodyCompletionsEnabled && !context.newCodyCompletionsEnabled) {
              Project[] openProjects = ProjectManager.getInstance().getOpenProjects();
              CodyCompletionsManager codyCompletionsManager = CodyCompletionsManager.getInstance();
              Arrays.stream(openProjects)
                  .flatMap(
                      project ->
                          Arrays.stream(FileEditorManager.getInstance(project).getAllEditors()))
                  .filter(fileEditor -> fileEditor instanceof TextEditor)
                  .map(fileEditor -> ((TextEditor) fileEditor).getEditor())
                  .forEach(codyCompletionsManager::clearCompletions);
            }
          }
        });
  }

  private void notifyAboutSuccessfulConnection() {
    KeyboardShortcut altSShortcut =
        new KeyboardShortcut(KeyStroke.getKeyStroke(KeyEvent.VK_S, InputEvent.ALT_DOWN_MASK), null);
    String altSShortcutText = KeymapUtil.getShortcutText(altSShortcut);
    Notification notification =
        new Notification(
            "Sourcegraph access",
            "Sourcegraph authentication success",
            "Your Sourcegraph account has been connected to the Sourcegraph plugin. Press "
                + altSShortcutText
                + " to open Sourcegraph.",
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

  public void setJavaToJSBridge(JavaToJSBridge javaToJSBridge) {
    this.javaToJSBridge = javaToJSBridge;
  }

  @Override
  public void dispose() {
    connection.disconnect();
  }
}
