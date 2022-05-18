package com.sourcegraph.browser;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.intellij.ui.jcef.JBCefJSQuery;
import com.sourcegraph.config.ThemeUtil;
import com.sourcegraph.find.PreviewContent;
import com.sourcegraph.find.PreviewPanel;
import org.jetbrains.annotations.Nullable;

public class JSToJavaBridgeRequestHandler {
    private final PreviewPanel previewPanel;

    public JSToJavaBridgeRequestHandler(PreviewPanel previewPanel) {
        this.previewPanel = previewPanel;
    }

    public JBCefJSQuery.Response handle(JsonObject request) {
        String action = request.get("action").getAsString();
        JsonObject arguments = request.getAsJsonObject("arguments");
        Gson gson = new Gson();
        PreviewContent previewContent;
        switch (action) {
            case "getTheme":
                JsonObject currentThemeAsJson = ThemeUtil.getCurrentThemeAsJson();
                return createResponse(currentThemeAsJson);
            case "preview":
                previewContent = gson.fromJson(arguments, PreviewContent.class);
                previewPanel.setContent(previewContent, false);
                return createResponse(null);
            case "clearPreview":
                previewPanel.clearContent();
                return createResponse(null);
            case "open":
                previewContent = gson.fromJson(arguments, PreviewContent.class);
                previewPanel.setContent(previewContent, true);
                return createResponse(null);
            default:
                return createResponse(2, "Unknown action: " + action);
        }
    }

    public JBCefJSQuery.Response handleInvalidRequest() {
        return createResponse(1, "Invalid JSON passed to bridge.");
    }

    private JBCefJSQuery.Response createResponse(@Nullable JsonObject result) {
        return new JBCefJSQuery.Response(result != null ? result.toString() : null);
    }

    private JBCefJSQuery.Response createResponse(int errorCode, @Nullable String errorMessage) {
        return new JBCefJSQuery.Response(null, errorCode, errorMessage);
    }
}

