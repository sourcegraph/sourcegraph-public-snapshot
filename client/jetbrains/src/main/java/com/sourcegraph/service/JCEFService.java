package com.sourcegraph.service;

import com.intellij.openapi.project.Project;
import com.sourcegraph.ui.JCEFWindow;
import com.sourcegraph.ui.SourcegraphWindow;

import java.util.Objects;

public class JCEFService {
    private final SourcegraphWindow sourcegraphWindow;
    private final JCEFWindow jcefWindow;

    public JCEFService(Project project) {
        jcefWindow = new JCEFWindow(project);
        sourcegraphWindow = new SourcegraphWindow(Objects.requireNonNull(project));
    }

    public JCEFWindow getJcefWindow() {
        return jcefWindow;
    }

    public SourcegraphWindow getSourcegraphWindow() {
        return this.sourcegraphWindow;
    }
}
