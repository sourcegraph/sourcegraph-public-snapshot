package com.sourcegraph.api;

import com.google.gson.JsonObject;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.utils.URIBuilder;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.nio.charset.StandardCharsets;

public class GraphQl {
    public static int callGraphQLService(@NotNull String instanceUrl, @Nullable String accessToken, @NotNull String query, @NotNull JsonObject variables) throws IOException {
        HttpPost request = createRequest(instanceUrl, accessToken, query, variables);
        try (CloseableHttpClient client = HttpClientBuilder.create().build()) {
            CloseableHttpResponse response = client.execute(request);
            response.close();
            return response.getStatusLine().getStatusCode();
        }
    }

    @NotNull
    private static HttpPost createRequest(@NotNull String instanceUrl, @Nullable String accessToken, @NotNull String query, @NotNull JsonObject variables) {
        HttpPost request = new HttpPost(getGraphQLApiURI(instanceUrl));

        request.setHeader("Content-Type", "application/json");
        request.setHeader("X-Sourcegraph-Should-Trace", "false");
        if (accessToken != null) {
            request.setHeader("Authorization", "token " + accessToken);
        }

        JsonObject body = new JsonObject();
        body.addProperty("query", query);
        body.add("variables", variables);

        ContentType contentType = ContentType.create("application/json", StandardCharsets.UTF_8);

        StringEntity entity = new StringEntity(body.toString(), contentType);
        entity.setContentEncoding(StandardCharsets.UTF_8.toString());

        request.setEntity(entity);
        return request;
    }

    @NotNull
    private static URI getGraphQLApiURI(String instanceUrl) {
        try {
            return new URIBuilder(instanceUrl + ".api/graphql").build();
        } catch (URISyntaxException e) {
            throw new RuntimeException(e);
        }
    }
}
