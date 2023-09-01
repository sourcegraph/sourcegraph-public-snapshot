package com.sourcegraph.config;

import com.intellij.util.messages.Topic;
import org.jetbrains.annotations.NotNull;

public interface PluginSettingChangeActionNotifier {

  Topic<PluginSettingChangeActionNotifier> TOPIC =
      Topic.create(
          "Sourcegraph Cody + Code Search plugin settings have changed",
          PluginSettingChangeActionNotifier.class);

  void beforeAction(@NotNull PluginSettingChangeContext context);

  void afterAction(@NotNull PluginSettingChangeContext context);
}
