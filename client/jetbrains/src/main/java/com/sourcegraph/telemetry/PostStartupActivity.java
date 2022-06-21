package com.sourcegraph.telemetry;

import com.intellij.ide.plugins.IdeaPluginDescriptor;
import com.intellij.ide.plugins.PluginInstaller;
import com.intellij.ide.plugins.PluginStateListener;
import com.intellij.openapi.application.Application;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.startup.StartupActivity;
import com.sourcegraph.config.SourcegraphApplicationService;
import org.jetbrains.annotations.NotNull;

import java.util.UUID;

public class PostStartupActivity implements StartupActivity {
    @Override
    public void runActivity(@NotNull Project project) {
        Application app = ApplicationManager.getApplication();
        SourcegraphApplicationService applicationConfig = app.getService(SourcegraphApplicationService.class);

        // When no anonymous user ID is set yet, we create a new one and treat this as an installation event. This
        // likely means that the user has never started IntelliJ with our extension before
        if (applicationConfig.getAnonymousUserId() == null) {
            applicationConfig.anonymousUserId = generateAnonymousUserId();
        }

        PluginInstaller.addStateListener(new PluginStateListener() {
            public void install(@NotNull IdeaPluginDescriptor ideaPluginDescriptor) {
            }

            @Override
            public void uninstall(@NotNull IdeaPluginDescriptor ideaPluginDescriptor) {
                if (ideaPluginDescriptor.getPluginId().getIdString().equals("com.sourcegraph.jetbrains")) {
                    // Clear the anonymous user ID so that we can detect a new installation if the user re-enables the
                    // extension.
                    applicationConfig.anonymousUserId = null;
                }
            }
        });
    }

    private static String generateAnonymousUserId() {
        return UUID.randomUUID().toString();
    }
}
