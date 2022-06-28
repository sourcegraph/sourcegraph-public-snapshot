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
                if (javaToJSBridge != null) {
                    javaToJSBridge.callJS("pluginSettingsChanged", ConfigUtil.getConfigAsJson(project));
                }

                if (!Objects.equals(context.oldUrl, context.newUrl)) {
                    GraphQlLogger.logInstallEvent(project, ConfigUtil::setInstallEventLogged);
                } else if (!Objects.equals(context.oldAccessToken, context.newAccessToken) && !ConfigUtil.isInstallEventLogged()) {
                    GraphQlLogger.logInstallEvent(project, ConfigUtil::setInstallEventLogged);
                }
            }
        });
    }

    public void setJavaToJSBridge(JavaToJSBridge javaToJSBridge) {
        this.javaToJSBridge = javaToJSBridge;
    }

    @Override
    public void dispose() {
        connection.disconnect();
    }
}
