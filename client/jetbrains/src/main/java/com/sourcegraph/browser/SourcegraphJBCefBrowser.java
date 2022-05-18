package com.sourcegraph.browser;

import com.intellij.openapi.util.Disposer;
import com.intellij.ui.jcef.JBCefBrowser;
import com.sourcegraph.config.ThemeUtil;
import org.cef.CefApp;

public class SourcegraphJBCefBrowser extends JBCefBrowser {
    private final JSToJavaBridge jsToJavaBridge;

    public SourcegraphJBCefBrowser(JSToJavaBridgeRequestHandler requestHandler) {
        super("http://sourcegraph/html/index.html");
        // Create and set up JCEF browser
        CefApp.getInstance().registerSchemeHandlerFactory("http", "sourcegraph", new HttpSchemeHandlerFactory());
        this.setPageBackgroundColor(ThemeUtil.getPanelBackgroundColorHexString());

        // Create bridge, set up handlers, then run init function
        String initJSCode = "window.initializeSourcegraph(" + (ThemeUtil.isDarkTheme() ? "true" : "false") + ");";
        jsToJavaBridge = new JSToJavaBridge(this, requestHandler, initJSCode);
        Disposer.register(this, jsToJavaBridge);
    }

    public JSToJavaBridge getJsToJavaBridge() {
        return jsToJavaBridge;
    }

    public void focus() {
        this.getCefBrowser().setFocus(true);
    }
}
