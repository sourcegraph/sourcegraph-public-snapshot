package com.sourcegraph.telemetry;

import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginInstaller;
import com.intellij.ide.plugins.PluginStateListener;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.agent.CodyAgentManager;
import com.sourcegraph.cody.config.CodyApplicationSettings;
import com.sourcegraph.cody.config.notification.AccountSettingChangeListener;
import com.sourcegraph.cody.config.notification.CodySettingChangeListener;
import com.sourcegraph.cody.initialization.Activity;
import java.util.UUID;
import org.jetbrains.annotations.NotNull;

public class TelemetryInitializerActivity implements Activity {
  private static String generateAnonymousUserId() {
    return UUID.randomUUID().toString();
  }

  @Override
  public void runActivity(@NotNull Project project) {
    // Make sure that settings -ChangeListeners are loaded
    project.getService(AccountSettingChangeListener.class);
    project.getService(CodySettingChangeListener.class);

    // When no anonymous user ID is set yet, we create a new one and treat this as an installation
    // event.
    // This likely means that the user has never started IntelliJ with our extension before
    CodyApplicationSettings codyApplicationSettings = CodyApplicationSettings.getInstance();
    if (codyApplicationSettings.getAnonymousUserId() == null) {
      codyApplicationSettings.setAnonymousUserId(generateAnonymousUserId());
    }

    PluginInstaller.addStateListener(
        new PluginStateListener() {
          public void install(@NotNull IdeaPluginDescriptor ideaPluginDescriptor) {
            CodyAgentManager.startAgent(project);
            GraphQlLogger.logInstallEvent(project)
                .thenAccept(
                    wasSuccessful -> {
                      if (wasSuccessful) {
                        codyApplicationSettings.setInstallEventLogged(true);
                      }
                    });
          }

          @Override
          public void uninstall(@NotNull IdeaPluginDescriptor ideaPluginDescriptor) {
            CodyAgentManager.stopAgent(project);
            if (ideaPluginDescriptor
                .getPluginId()
                .getIdString()
                .equals("com.sourcegraph.jetbrains")) {
              GraphQlLogger.logUninstallEvent(project);

              // Clearing this so that we can detect a new installation if the user re-enables the
              // extension.
              codyApplicationSettings.setAnonymousUserId(null);
              codyApplicationSettings.setInstallEventLogged(false);
            }
          }
        });
  }
}
