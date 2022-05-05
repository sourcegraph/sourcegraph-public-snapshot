package com.sourcegraph.find;

import com.intellij.openapi.project.Project;

public class JCEFService {
    private final SourcegraphWindow sourcegraphWindow;

    public JCEFService(Project project) {
        sourcegraphWindow = new SourcegraphWindow(project);
    }

    public SourcegraphWindow getSourcegraphWindow() {
        return this.sourcegraphWindow;
    }
}
