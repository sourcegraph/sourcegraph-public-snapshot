package com.sourcegraph.config;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.util.messages.MessageBus;
import com.intellij.util.messages.MessageBusConnection;
import com.sourcegraph.browser.JavaToJSBridge;
import org.jetbrains.annotations.NotNull;

public class SettingsChangeListener implements Disposable {
    private final MessageBusConnection connection;

    public SettingsChangeListener(@NotNull Project project, @NotNull JavaToJSBridge javaToJSBridge) {
        MessageBus bus = project.getMessageBus();

        connection = bus.connect();
        connection.subscribe(PluginSettingChangeActionNotifier.TOPIC, new PluginSettingChangeActionNotifier() {
            @Override
            public void beforeAction(@Nullable String oldUrl, @Nullable String oldAccessToken, @Nullable String newUrl, @Nullable String newAccessToken) {
                // Do nothing
            }

            @Override
            public void afterAction(@Nullable String oldUrl, @Nullable String oldAccessToken, @Nullable String newUrl, @Nullable String newAccessToken) {
                javaToJSBridge.callJS("pluginSettingsChanged", ConfigUtil.getConfigAsJson(project));
            }
        });
    }

    @Override
    public void dispose() {
        connection.disconnect();
    }
}
