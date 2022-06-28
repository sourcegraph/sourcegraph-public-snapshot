package com.sourcegraph.browser;

import org.cef.browser.CefBrowser;
import org.cef.browser.CefFrame;
import org.cef.callback.CefSchemeHandlerFactory;
import org.cef.handler.CefResourceHandler;
import org.cef.network.CefRequest;

public class HttpSchemeHandlerFactory implements CefSchemeHandlerFactory {
    @Override
    public CefResourceHandler create(CefBrowser browser, CefFrame frame, String schemeName, CefRequest request) {
        return new HttpSchemeHandler();
    }
}

