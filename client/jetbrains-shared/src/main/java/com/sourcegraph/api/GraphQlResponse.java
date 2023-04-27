package com.sourcegraph.api;

import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import com.google.gson.JsonSyntaxException;
import org.jetbrains.annotations.NotNull;

public class GraphQlResponse {
    private final int statusCode;
    private final String body;

    public GraphQlResponse(int statusCode, String body) {
        this.statusCode = statusCode;
        this.body = body;
    }

    public int getStatusCode() {
        return statusCode;
    }

    public String getBody() {
        return body;
    }

    @NotNull
    public JsonObject getBodyAsJson() throws JsonSyntaxException {
        return JsonParser.parseString(body).getAsJsonObject();
    }

}
