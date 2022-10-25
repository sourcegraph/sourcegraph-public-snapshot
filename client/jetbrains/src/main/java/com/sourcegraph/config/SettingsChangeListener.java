package com.sourcegraph.config;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.KeyboardShortcut;
import com.intellij.openapi.keymap.KeymapUtil;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.util.messages.MessageBus;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.find.browser.JavaToJSBridge;
import com.sourcegraph.telemetry.GraphQlLogger;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.util.Objects;

public class SettingsChangeListener implements Disposable {
    private final MessageBusConnection connection;
    private JavaToJSBridge javaToJSBridge;

    public SettingsChangeListener(@NotNull Project project) {
        MessageBus bus = project.getMessageBus();

        connection = bus.connect();
        connection.subscribe(PluginSettingChangeActionNotifier.TOPIC, new PluginSettingChangeActionNotifier() {
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
                } else if (!Objects.equals(context.oldAccessToken, context.newAccessToken) && !ConfigUtil.isInstallEventLogged()) {
                    GraphQlLogger.logInstallEvent(project, ConfigUtil::setInstallEventLogged);
                }

                // Notify user about a successful connection
                if (context.newUrl != null) {
                    ApiAuthenticator.testConnection(context.newUrl, context.newAccessToken, context.newCustomRequestHeaders, (status) -> {
                        if (ConfigUtil.didAuthenticationFailLastTime() && status == ApiAuthenticator.ConnectionStatus.AUTHENTICATED) {
                            notifyAboutSuccessfulConnection();
                        }
                        ConfigUtil.setAuthenticationFailedLastTime(status != ApiAuthenticator.ConnectionStatus.AUTHENTICATED);
                    });
                }
            }
        });
    }

    private void notifyAboutSuccessfulConnection() {
        KeyboardShortcut altSShortcut = new KeyboardShortcut(KeyStroke.getKeyStroke(KeyEvent.VK_S, InputEvent.ALT_DOWN_MASK), null);
        String altSShortcutText = KeymapUtil.getShortcutText(altSShortcut);
        Notification notification = new Notification("Sourcegraph access", "Sourcegraph authentication success",
            "Your Sourcegraph account has been connected to the Sourcegraph plugin. Press " + altSShortcutText + " to open Sourcegraph.", NotificationType.INFORMATION);
        AnAction dismissAction = new DumbAwareAction("Dismiss") {
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
