package com.sourcegraph.ui;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.ui.jcef.JBCefApp;
import com.intellij.ui.jcef.JBCefBrowser;
import com.intellij.ui.jcef.JBCefBrowserBase;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.scheme.SchemeHandlerFactory;
import org.cef.CefApp;
import org.cef.browser.CefBrowser;

import javax.swing.*;
import java.awt.*;

public class JCEFWindow {
    private JPanel panel;
    private Project project;
    private JBCefBrowserBase browser;
    private CefBrowser cefBrowser;

    public JCEFWindow(Project project) {
        this.project = project;
        panel = new JPanel(new BorderLayout());

        if (!JBCefApp.isSupported()) {
            JLabel warningLabel = new JLabel("Unfortunately, the browser is not available on your system. Try running the IDE with the default OpenJDK.");
            panel.add(warningLabel);
            return;
        }

        this.browser = new JBCefBrowser("http://sourcegraph/html/index.html") ;
        this.cefBrowser= browser.getCefBrowser();

        CefApp
            .getInstance()
            .registerSchemeHandlerFactory(
    "http",
    "sourcegraph",
                new SchemeHandlerFactory()
            );

        panel.add(this.browser.getComponent(),BorderLayout.CENTER);


        String backgroundColor = "#" + Integer.toHexString(UIUtil.getPanelBackground().getRGB()).substring(2);
        this.browser.setPageBackgroundColor(backgroundColor);

        Disposer.register(project, this.browser);
    }

    public JPanel getContent() {
        return panel;
    }

    public void focus() {
        this.cefBrowser.setFocus(true);
    }
}
