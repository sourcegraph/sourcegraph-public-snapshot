package com.sourcegraph.service;

import com.intellij.openapi.project.Project;
import com.sourcegraph.ui.SourcegraphJBCefBrowser;
import com.sourcegraph.ui.SourcegraphWindow;

public class JCEFService {
    private final SourcegraphWindow sourcegraphWindow;
    private final SourcegraphJBCefBrowser sourcegraphJBCefBrowser;

    public JCEFService(Project project) {
        sourcegraphJBCefBrowser = new SourcegraphJBCefBrowser();
        sourcegraphWindow = new SourcegraphWindow(project, this);
    }

    public SourcegraphJBCefBrowser getJcefWindow() {
        return sourcegraphJBCefBrowser;
    }

    public SourcegraphWindow getSourcegraphWindow() {
        return this.sourcegraphWindow;
    }
}
