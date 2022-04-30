package com.sourcegraph.service;

import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.google.gson.JsonSyntaxException;
import com.intellij.openapi.Disposable;
import com.intellij.ui.jcef.JBCefBrowserBase;
import com.intellij.ui.jcef.JBCefJSQuery;
import org.cef.browser.CefBrowser;
import org.cef.browser.CefFrame;
import org.cef.handler.CefLoadHandler;
import org.cef.network.CefRequest;

public class JSToJavaBridge implements Disposable {
    JBCefJSQuery jsToJavaBridge;

    public JSToJavaBridge(JBCefBrowserBase browser,
                          JSToJavaBridgeRequestHandler requestHandler,
                          String jsCodeToRunAfterBridgeInit) {
        jsToJavaBridge = JBCefJSQuery.create(browser);
        jsToJavaBridge.addHandler((String requestAsString) -> {
            try {
                JsonObject requestAsJson = JsonParser.parseString(requestAsString).getAsJsonObject();
                return requestHandler.handle(requestAsJson);
            } catch (JsonSyntaxException e) {
                return requestHandler.handleInvalidRequest();
            }
        });

        browser.getJBCefClient().addLoadHandler(new CefLoadHandler() {
            @Override
            public void onLoadingStateChange(CefBrowser cefBrowser, boolean isLoading, boolean canGoBack, boolean canGoForward) {
            }

            @Override
            public void onLoadStart(CefBrowser cefBrowser, CefFrame frame, CefRequest.TransitionType transitionType) {
            }

            @Override
            public void onLoadEnd(CefBrowser cefBrowser, CefFrame frame, int httpStatusCode) {
                cefBrowser.executeJavaScript(
                    "window.javaBridge = async function(request) {" +
                        "    return new Promise((resolve, reject) => { " +
                        jsToJavaBridge.inject("request", "resolve", "reject") +
                        "    });" +
                        "};",
                    cefBrowser.getURL(), 0);
                cefBrowser.executeJavaScript(jsCodeToRunAfterBridgeInit, "", 0);
            }

            @Override
            public void onLoadError(CefBrowser cefBrowser, CefFrame frame, ErrorCode errorCode, String errorText, String failedUrl) {
            }
        }, browser.getCefBrowser());

    }

    @Override
    public void dispose() {
        jsToJavaBridge.dispose();
    }
}
