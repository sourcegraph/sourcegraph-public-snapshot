package com.sourcegraph.ui;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.ui.jcef.JBCefBrowser;
import com.intellij.ui.jcef.JBCefBrowserBase;
import com.sourcegraph.bridge.JSToJavaBridge;
import com.sourcegraph.bridge.JSToJavaBridgeRequestHandler;
import com.sourcegraph.scheme.SchemeHandlerFactory;
import org.cef.CefApp;
import org.cef.browser.CefBrowser;

import javax.swing.*;
import java.awt.*;
import java.util.Objects;

public class JCEFWindow {
    private final JPanel panel;
    private CefBrowser cefBrowser;

    public JCEFWindow(Project project) {
        panel = new JPanel(new BorderLayout());

        /* Make sure JCEF is supported */
        if (!JBCefApp.isSupported()) {
            JLabel warningLabel = new JLabel("Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");
            panel.add(warningLabel);
            return;
        }

        /* Create and set up JCEF browser */
        JBCefBrowserBase browser = new JBCefBrowser("http://sourcegraph/html/index.html");
        this.cefBrowser = browser.getCefBrowser();
        CefApp.getInstance().registerSchemeHandlerFactory("http", "sourcegraph", new SchemeHandlerFactory());
        browser.setPageBackgroundColor(ThemeService.getPanelBackgroundColorHexString());
        Disposer.register(project, browser);

        /* Add browser to panel */
        panel.add(Objects.requireNonNull(browser.getComponent()), BorderLayout.CENTER);

        /* Create bridge, set up handlers, then run init function */
        String initJSCode = "window.initializeSourcegraph(" + (ThemeService.isDarkTheme() ? "true" : "false") + ");";
        JSToJavaBridge bridge = new JSToJavaBridge(browser, new JSToJavaBridgeRequestHandler(), initJSCode);
        Disposer.register(browser, bridge);
    }

    public JPanel getContent() {
        return panel;
    }

    public void focus() {
        this.cefBrowser.setFocus(true);
    }
}
