package com.sourcegraph.find;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;

public class JCEFService implements Disposable {
    private final SourcegraphWindow sourcegraphWindow;

    public JCEFService(Project project) {
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
