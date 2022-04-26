package com.sourcegraph.service;

import com.intellij.openapi.project.Project;
import com.sourcegraph.ui.JCEFWindow;

public class JCEFService {
    private Project project;
    private JCEFWindow window;
    public JCEFService(Project project) {
        this.project = project;
        this.window = new JCEFWindow(project);
    }

    public JCEFWindow getWindow() {
        return this.window;
    }
}
