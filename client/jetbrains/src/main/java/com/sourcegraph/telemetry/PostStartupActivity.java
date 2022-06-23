package com.sourcegraph.telemetry;

import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginInstaller;
import com.intellij.ide.plugins.PluginStateListener;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.startup.StartupActivity;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;

public class PostStartupActivity implements StartupActivity {
    @Override
    public void runActivity(@NotNull Project project) {
        // When no anonymous user ID is set yet, we create a new one and treat this as an installation event.
        // This likely means that the user has never started IntelliJ with our extension before
        if (ConfigUtil.getAnonymousUserId() == null) {
            ConfigUtil.generateAndSetAnonymousUserId();
        }

        PluginInstaller.addStateListener(new PluginStateListener() {
            public void install(@NotNull IdeaPluginDescriptor ideaPluginDescriptor) {
                GraphQlLogger.logInstallEvent(project);
            }

            @Override
            public void uninstall(@NotNull IdeaPluginDescriptor ideaPluginDescriptor) {
                if (ideaPluginDescriptor.getPluginId().getIdString().equals("com.sourcegraph.jetbrains")) {
                    GraphQlLogger.logUninstallEvent(project);

                    // Clearing this so that we can detect a new installation if the user re-enables the extension.
                    ConfigUtil.clearAnonymousUserId();
                }
            }
        });
    }
}
