package com.sourcegraph.cody;

import com.intellij.ide.plugins.PluginManagerCore;
import com.intellij.openapi.extensions.PluginId;
import org.jetbrains.annotations.NotNull;

public class PluginUtil {

  /**
   * List of known plugin IDs you want to check. You can obtain the plugin IDs from each plugin's
   * plugin.xml file, usually under the <id> tag. Add as many as you need.
   */
  private static final String[] KNOWN_PLUGINS = {
    "com.github.copilot",
    "com.tabnine.TabNine",
    "com.tabnine.TabNine-Enterprise",
    "aws.toolkit", // Includes CodeWhisperer
    "com.codeium.intellij",
    "com.codeium.enterpriseUpdater",
    "com.nnthink.aixcoder",
    "com.codota.csp.intellij",
    "com.github.simiacryptus.intellijopenaicodeassist",
    "com.tabbyml.intellij-tabby",
  };

  public static boolean isAnyKnownPluginEnabled() {
    for (String pluginId : KNOWN_PLUGINS) {
      if (isPluginInstalledAndEnabled(PluginId.getId(pluginId))) {
        return true;
      }
    }
    return false;
  }

  private static boolean isPluginInstalledAndEnabled(@NotNull PluginId pluginId) {
    return PluginManagerCore.isPluginInstalled(pluginId) && !PluginManagerCore.isDisabled(pluginId);
  }
}
