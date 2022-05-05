package com.sourcegraph.find;

import com.intellij.openapi.project.Project;
import com.sourcegraph.browser.SourcegraphJBCefBrowser;

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
