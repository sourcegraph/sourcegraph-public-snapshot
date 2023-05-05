package com.sourcegraph.completions;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import org.jetbrains.annotations.NotNull;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class CompletionsService {
    private final String instanceUrl;
    private final String token;

    public CompletionsService(@NotNull String instanceUrl, @NotNull String token) {
        this.instanceUrl = instanceUrl;
        this.token = token;
    }

    /**
     * Sends a completions request to the Sourcegraph instance, and returns the response.
     */
    public String getCompletion(@NotNull CompletionsInput input) throws IOException, InterruptedException {
        Gson gson = new Gson();

        String query = "query completions($input: CompletionsInput!) { completions(input: $input) }";
        var wrapper = new GraphQLWrapper(query).withVariable("input", input);
        if (wrapper == null) {
            return null;
        }
        var json = wrapper.toJsonString();

        URI uri = URI.create(instanceUrl);
        var body = HttpClient.newHttpClient().send(HttpRequest.newBuilder(uri)
                .POST(HttpRequest.BodyPublishers.ofString(json))
                .header("Authorization", "token " + token)
                .build(), HttpResponse.BodyHandlers.ofString())
            .body();
        if (body == null) {
            return null;
        }

        return gson.fromJson(body, JsonObject.class)
            .getAsJsonObject("data")
            .getAsJsonPrimitive("completions")
            .getAsString();
    }
}
