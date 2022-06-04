package com.sourcegraph.find;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import org.jetbrains.annotations.NotNull;

public class JCEFService implements Disposable {
    private final SourcegraphWindow sourcegraphWindow;

    public JCEFService(@NotNull Project project) {
        sourcegraphWindow = new SourcegraphWindow(project);
    }

    public SourcegraphWindow getSourcegraphWindow() {
        return this.sourcegraphWindow;
    }

    @Override
    public void dispose() {
        sourcegraphWindow.dispose();
    }
}
