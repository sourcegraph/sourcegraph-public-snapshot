package com.sourcegraph.website;

import com.intellij.openapi.project.Project;
import com.sourcegraph.common.BrowserOpener;
import org.jetbrains.annotations.NotNull;

public class OpenFileAction extends FileActionBase {
    @Override
    protected void handleFileUri(@NotNull Project project, @NotNull String url) {
        BrowserOpener.openInBrowser(project, url);
    }
}
