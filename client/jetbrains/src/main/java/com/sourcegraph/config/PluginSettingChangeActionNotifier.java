package com.sourcegraph.config;

import com.intellij.util.messages.Topic;
import org.jetbrains.annotations.Nullable;

public interface PluginSettingChangeActionNotifier {

    Topic<PluginSettingChangeActionNotifier> TOPIC = Topic.create("Sourcegraph plugin settings have changed", PluginSettingChangeActionNotifier.class);

    void beforeAction(@Nullable String oldUrl, @Nullable String oldAccessToken, @Nullable String newUrl, @Nullable String newAccessToken);

    void afterAction(@Nullable String oldUrl, @Nullable String oldAccessToken, @Nullable String newUrl, @Nullable String newAccessToken);
}
