package com.sourcegraph.cody.completions;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.sourcegraph.api.GraphQlClient;
import org.jetbrains.annotations.NotNull;

import java.io.IOException;

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
        var variables = new JsonObject();
        variables.add("input", gson.toJsonTree(input));

        var response = GraphQlClient.callGraphQLService(instanceUrl, token, null, query, variables);
        return response
            .getBodyAsJson()
            .getAsJsonObject("data")
            .getAsJsonPrimitive("completions")
            .getAsString();
    }
}
