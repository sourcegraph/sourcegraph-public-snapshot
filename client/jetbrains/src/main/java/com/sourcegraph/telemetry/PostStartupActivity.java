package com.sourcegraph.telemetry;

import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginInstaller;
import com.intellij.ide.plugins.PluginStateListener;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.startup.StartupActivity;
import com.sourcegraph.cody.CodyAgentProjectListener;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.SettingsChangeListener;
import java.util.UUID;
import org.jetbrains.annotations.NotNull;

public class PostStartupActivity implements StartupActivity.DumbAware {
  private static String generateAnonymousUserId() {
    return UUID.randomUUID().toString();
  }

  @Override
  public void runActivity(@NotNull Project project) {
    // Make sure that SettingsChangeListener is loaded
    project.getService(SettingsChangeListener.class);

    // When no anonymous user ID is set yet, we create a new one and treat this as an installation
    // event.
    // This likely means that the user has never started IntelliJ with our extension before
    if (ConfigUtil.getAnonymousUserId() == null) {
      ConfigUtil.setAnonymousUserId(generateAnonymousUserId());
    }

    PluginInstaller.addStateListener(
        new PluginStateListener() {
          public void install(@NotNull IdeaPluginDescriptor ideaPluginDescriptor) {
            CodyAgentProjectListener.startAgent(project);
            GraphQlLogger.logInstallEvent(project)
                .thenAccept(
                    wasSuccessful -> {
                      if (wasSuccessful) {
                        ConfigUtil.setInstallEventLogged(true);
                      }
                    });
          }

          @Override
          public void uninstall(@NotNull IdeaPluginDescriptor ideaPluginDescriptor) {
            CodyAgentProjectListener.stopAgent(project);
            if (ideaPluginDescriptor
                .getPluginId()
                .getIdString()
                .equals("com.sourcegraph.jetbrains")) {
              GraphQlLogger.logUninstallEvent(project);

              // Clearing this so that we can detect a new installation if the user re-enables the
              // extension.
              ConfigUtil.setAnonymousUserId(null);

              ConfigUtil.setInstallEventLogged(false);
            }
          }
        });
  }
}
