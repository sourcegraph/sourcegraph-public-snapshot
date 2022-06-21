package com.sourcegraph.config;

import com.intellij.util.messages.Topic;

public interface PluginSettingChangeActionNotifier {

    Topic<PluginSettingChangeActionNotifier> TOPIC = Topic.create("Sourcegraph plugin settings have changed", PluginSettingChangeActionNotifier.class);

    void beforeAction();

    void afterAction();
}
