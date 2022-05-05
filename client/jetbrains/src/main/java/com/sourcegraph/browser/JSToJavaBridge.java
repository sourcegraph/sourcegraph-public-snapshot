package com.sourcegraph.browser;

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
    JBCefJSQuery query;

    public JSToJavaBridge(JBCefBrowserBase browser,
                          JSToJavaBridgeRequestHandler requestHandler,
                          String jsCodeToRunAfterBridgeInit) {
        query = JBCefJSQuery.create(browser);
        query.addHandler((String requestAsString) -> {
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
                // In case of a failure, Java returns two arguments, so must use an intermediate function.
                // (source: https://dploeger.github.io/intellij-api-doc/com/intellij/ui/jcef/JBCefJSQuery.html#:~:text=onFailureCallback%20%2D%20JS%20callback%20in%20format%3A%20function(error_code%2C%20error_message)%20%7B%7D)
                cefBrowser.executeJavaScript(
                    "window.callJava = function(request) {" +
                        "    return new Promise((resolve, reject) => { " +
                        "        const requestAsString = JSON.stringify(request);" +
                        "        const onSuccessCallback = responseAsString => {" +
                        "            resolve(JSON.parse(responseAsString));" +
                        "        };" +
                        "        const onFailureCallback = (errorCode, errorMessage) => {" +
                        "            reject(new Error(`${errorCode} - ${errorMessage}`));" +
                        "        };" +
                        "        " + query.inject("requestAsString", "onSuccessCallback", "onFailureCallback") +
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
        query.dispose();
    }
}
