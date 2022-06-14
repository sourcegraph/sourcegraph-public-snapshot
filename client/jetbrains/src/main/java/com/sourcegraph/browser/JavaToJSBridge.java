package com.sourcegraph.browser;

import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.ui.jcef.JBCefBrowserBase;
import com.intellij.ui.jcef.JBCefJSQuery;

import java.util.function.Consumer;
import java.util.function.Function;

public class JavaToJSBridge {
    private final int QUERY_COUNT = 2;
    private final JBCefBrowserBase browser;
    private final Logger logger = Logger.getInstance(JavaToJSBridge.class);
    private final JBCefJSQuery[] queries = new JBCefJSQuery[QUERY_COUNT];
    private final boolean[] isQueryRunning = new boolean[QUERY_COUNT];
    private Function<String, JBCefJSQuery.Response> handler = null;

    public JavaToJSBridge(JBCefBrowserBase browser) {
        this.browser = browser;
        this.queries[0] = JBCefJSQuery.create(browser);
        this.isQueryRunning[0] = false;
        this.queries[1] = JBCefJSQuery.create(browser);
        this.isQueryRunning[1] = false;
    }

    public void callJS(String action, JsonObject arguments) {
        this.callJS(action, arguments, null);
    }

    public void callJS(String action, JsonObject arguments, Consumer<JsonObject> callback) {
        // Reason for the locking:
        // JBCefJSQuery objects MUST be created before the browser is loaded, otherwise an error is thrown.
        // As there is only one JBCefJSQuery object, and we need to wait for the result of the last execution,
        // we can only run one query at a time.
        // If this becomes a bottleneck, we can create a pool of JBCefJSQuery objects to bridge this,
        // or find a different solution.
        if (hasAvailableQuery()) {
            int queryIndex = isQueryRunning[0] ? 1 : 0;
            isQueryRunning[queryIndex] = true;
            JBCefJSQuery query = queries[queryIndex];
            String js = "window.callJS('" + action + "', '" + (arguments != null ? arguments.toString() : "null") + "', (result) => {" +
                "    " + query.inject("result") +
                "});";

            handler = responseAsString -> {
                if (callback != null) {
                    callback.accept(JsonParser.parseString(responseAsString).getAsJsonObject());
                }
                query.removeHandler(handler);
                handler = null;
                isQueryRunning[queryIndex] = false;
                return null;
            };
            query.addHandler(handler);
            browser.getCefBrowser().executeJavaScript(js, browser.getCefBrowser().getURL(), 0);
        } else {
            logger.error("Query is already running, ignoring callJS");
        }
    }

    public boolean hasAvailableQuery() {
        return !(isQueryRunning[0] && isQueryRunning[1]);
    }
}
