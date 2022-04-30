package com.sourcegraph.service;

import com.google.gson.JsonObject;
import com.intellij.ui.jcef.JBCefJSQuery;

public class JSToJavaBridgeRequestHandler {
    public JBCefJSQuery.Response handle(JsonObject request) {
        String action = request.get("action").getAsString();
        // JsonObject arguments = request.getAsJsonObject("arguments");
        return createResponse(false, "Unknown action: " + action, null);
    }

    public JBCefJSQuery.Response handleInvalidRequest() {
        return createResponse(false, "Invalid JSON passed to bridge.", null);
    }

    private JBCefJSQuery.Response createResponse(boolean success, String errorMessage, JsonObject data) {
        JsonObject response = new JsonObject();
        response.addProperty("success", success);
        response.addProperty("errorMessage", errorMessage);
        response.add("data", data);
        return new JBCefJSQuery.Response(response.getAsString());
    }
}
