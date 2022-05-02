package com.sourcegraph.bridge;

import com.intellij.ui.jcef.JBCefBrowserBase;
import com.intellij.ui.jcef.JBCefJSQuery;

public class JavaToJSBridge {
    private final JBCefBrowserBase browser;

    public JavaToJSBridge(JBCefBrowserBase browser) {
        this.browser = browser;
    }

    public void callJS(String action, String data) {
        JBCefJSQuery query = JBCefJSQuery.create(browser);
        String js = "window.callJS('" + action + "', '" + data + "', callback: (result) => {" +
            "    " + query.inject("result") + ";" +
            "});";
        query.addHandler((String responseAsString) -> {
            System.out.println("Response: " + responseAsString);
            return null;
        });
        browser.getCefBrowser().executeJavaScript(js, browser.getCefBrowser().getURL(), 0);
    }
}
