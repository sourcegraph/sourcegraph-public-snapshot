package com.sourcegraph.find.browser;

import com.intellij.openapi.util.Disposer;
import com.intellij.ui.jcef.JBCefBrowser;
import com.sourcegraph.config.SettingsChangeListener;
import com.sourcegraph.config.ThemeUtil;
import javax.swing.*;
import org.cef.CefApp;
import org.jetbrains.annotations.NotNull;

public class SourcegraphJBCefBrowser extends JBCefBrowser {
  private final JavaToJSBridge javaToJSBridge;

  public SourcegraphJBCefBrowser(@NotNull JSToJavaBridgeRequestHandler requestHandler) {
    super("http://sourcegraph/html/index.html");
    // Create and set up JCEF browser
    CefApp.getInstance()
        .registerSchemeHandlerFactory("http", "sourcegraph", new HttpSchemeHandlerFactory());

    // Create bridges, set up handlers, then run init function
    String initJSCode = "window.initializeSourcegraph();";
    JSToJavaBridge jsToJavaBridge = new JSToJavaBridge(this, requestHandler, initJSCode);
    Disposer.register(this, jsToJavaBridge);
    javaToJSBridge = new JavaToJSBridge(this);

    requestHandler
        .getProject()
        .getService(SettingsChangeListener.class)
        .setJavaToJSBridge(javaToJSBridge);

    UIManager.addPropertyChangeListener(
        propertyChangeEvent -> {
          if (propertyChangeEvent.getPropertyName().equals("lookAndFeel")) {
            javaToJSBridge.callJS("themeChanged", ThemeUtil.getCurrentThemeAsJson());
          }
        });
  }

  @NotNull
  public JavaToJSBridge getJavaToJSBridge() {
    return javaToJSBridge;
  }
}
