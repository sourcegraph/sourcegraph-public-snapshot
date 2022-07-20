package com.sourcegraph.website;

import com.intellij.openapi.project.Project;
import com.sourcegraph.common.BrowserErrorNotification;
import org.jetbrains.annotations.NotNull;

import java.awt.*;
import java.io.IOException;
import java.net.URI;

public class OpenFileAction extends FileActionBase {
    @Override
    protected void handleFileUri(@NotNull Project project, @NotNull String url) {
        URI uri = URI.create(url);

        try {
            Desktop.getDesktop().browse(uri);
        } catch (IOException | UnsupportedOperationException e) {
            BrowserErrorNotification.show(project, uri);
        }
    }
}
