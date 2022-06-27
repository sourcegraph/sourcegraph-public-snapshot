package com.sourcegraph.config;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.util.messages.MessageBus;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.browser.JavaToJSBridge;
import com.sourcegraph.telemetry.GraphQlLogger;
import org.jetbrains.annotations.NotNull;

import java.util.Objects;

public class SettingsChangeListener implements Disposable {
    private final MessageBusConnection connection;

    public SettingsChangeListener(@NotNull Project project, @NotNull JavaToJSBridge javaToJSBridge) {
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
                javaToJSBridge.callJS("pluginSettingsChanged", ConfigUtil.getConfigAsJson(project));

                if (!Objects.equals(context.oldUrl, context.newUrl)) {
                    GraphQlLogger.logInstallEvent(project, (wasSuccessful) -> {
                        if (wasSuccessful) {
                            ConfigUtil.setInstallEventLogged(true);
                        }
                    });
                } else if (!Objects.equals(context.oldAccessToken, context.newAccessToken) && !ConfigUtil.isInstallEventLogged()) {
                    GraphQlLogger.logInstallEvent(project, (wasSuccessful) -> {
                        if (wasSuccessful) {
                            ConfigUtil.setInstallEventLogged(true);
                        }
                    });
                }
            }
        });
    }

    @Override
    public void dispose() {
        connection.disconnect();
    }
}
