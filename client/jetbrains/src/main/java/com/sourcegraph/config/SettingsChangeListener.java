package com.sourcegraph.config;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.util.messages.MessageBus;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.browser.JavaToJSBridge;
import com.sourcegraph.telemetry.GraphQlLogger;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.Objects;

public class SettingsChangeListener implements Disposable {
    private final MessageBusConnection connection;

    public SettingsChangeListener(@NotNull Project project, @NotNull JavaToJSBridge javaToJSBridge) {
        MessageBus bus = project.getMessageBus();

        connection = bus.connect();
        connection.subscribe(PluginSettingChangeActionNotifier.TOPIC, new PluginSettingChangeActionNotifier() {
            @Override
            public void beforeAction(@Nullable String oldUrl, @Nullable String oldAccessToken, @Nullable String newUrl, @Nullable String newAccessToken) {
                if (!Objects.equals(oldUrl, newUrl)) {
                    GraphQlLogger.logUninstallEvent(project);
                    ConfigUtil.setInstallEventLogged(false);
                }
            }

            @Override
            public void afterAction(@Nullable String oldUrl, @Nullable String oldAccessToken, @Nullable String newUrl, @Nullable String newAccessToken) {
                javaToJSBridge.callJS("pluginSettingsChanged", ConfigUtil.getConfigAsJson(project));

                if (!Objects.equals(oldUrl, newUrl)) {
                    GraphQlLogger.logInstallEvent(project, (wasSuccessful) -> {
                        if (wasSuccessful) {
                            ConfigUtil.setInstallEventLogged(true);
                        }
                    });
                } else if (!Objects.equals(oldAccessToken, newAccessToken) && !ConfigUtil.isInstallEventLogged()) {
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
