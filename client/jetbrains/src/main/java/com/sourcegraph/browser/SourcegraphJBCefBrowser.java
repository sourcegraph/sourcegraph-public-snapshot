package com.sourcegraph.browser;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Disposer;
import com.intellij.ui.jcef.JBCefBrowser;
import com.sourcegraph.config.SettingsChangeListener;
import com.sourcegraph.config.ThemeUtil;
import org.cef.CefApp;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;

public class SourcegraphJBCefBrowser extends JBCefBrowser {

    private final SettingsChangeListener settingsChangeListener;

    public SourcegraphJBCefBrowser(@NotNull JSToJavaBridgeRequestHandler requestHandler) {
        super("http://sourcegraph/html/index.html");
        // Create and set up JCEF browser
        CefApp.getInstance().registerSchemeHandlerFactory("http", "sourcegraph", new HttpSchemeHandlerFactory());
        this.setPageBackgroundColor(ThemeUtil.getPanelBackgroundColorHexString());

        // Create bridges, set up handlers, then run init function
        String initJSCode = "window.initializeSourcegraph();";
        JSToJavaBridge jsToJavaBridge = new JSToJavaBridge(this, requestHandler, initJSCode);
        Disposer.register(this, jsToJavaBridge);
        JavaToJSBridge javaToJSBridge = new JavaToJSBridge(this);

        Project project = requestHandler.getProject();
        settingsChangeListener = new SettingsChangeListener(project, javaToJSBridge);

        UIManager.addPropertyChangeListener(propertyChangeEvent -> {
            if (propertyChangeEvent.getPropertyName().equals("lookAndFeel")) {
                javaToJSBridge.callJS("themeChanged", ThemeUtil.getCurrentThemeAsJson());
            }
        });
    }

    public void focus() {
        this.getCefBrowser().setFocus(true);
    }

    @Override
    public void dispose() {
        super.dispose();

        settingsChangeListener.dispose();
    }
}
