package com.sourcegraph.ui;

import com.intellij.ui.jcef.JBCefBrowser;
import com.sourcegraph.bridge.JSToJavaBridge;
import com.sourcegraph.bridge.JSToJavaBridgeRequestHandler;
import com.sourcegraph.scheme.HttpSchemeHandlerFactory;
import org.cef.CefApp;

public class SourcegraphJBCefBrowser extends JBCefBrowser {
    private final JSToJavaBridge jsToJavaBridge;

    public SourcegraphJBCefBrowser() {
        super("http://sourcegraph/html/index.html");
        /* Create and set up JCEF browser */
        CefApp.getInstance().registerSchemeHandlerFactory("http", "sourcegraph", new HttpSchemeHandlerFactory());
        this.setPageBackgroundColor(ThemeService.getPanelBackgroundColorHexString());

        /* Create bridge, set up handlers, then run init function */
        String initJSCode = "window.initializeSourcegraph(" + (ThemeService.isDarkTheme() ? "true" : "false") + ");";
        jsToJavaBridge = new JSToJavaBridge(this, new JSToJavaBridgeRequestHandler(), initJSCode);
    }

    public JSToJavaBridge getJsToJavaBridge() {
        return jsToJavaBridge;
    }

    public void focus() {
        this.getCefBrowser().setFocus(true);
    }
}
