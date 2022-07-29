package com.sourcegraph.api;

import com.google.gson.JsonObject;
import com.google.gson.JsonSyntaxException;
import org.apache.http.HttpEntity;
import org.apache.http.HttpResponse;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.utils.URIBuilder;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.util.EntityUtils;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.nio.charset.StandardCharsets;

public class GraphQlClient {
    @NotNull
    public static GraphQlResponse callGraphQLService(@NotNull String instanceUrl, @Nullable String accessToken, @NotNull String query, @NotNull JsonObject variables) throws IOException {
        HttpPost request = createRequest(instanceUrl, accessToken, query, variables);
        try (CloseableHttpClient client = HttpClientBuilder.create().build()) {
            CloseableHttpResponse httpResponse = client.execute(request);
            GraphQlResponse response = new GraphQlResponse(getStatusCode(httpResponse), getResponseBody(httpResponse));
            httpResponse.close();
            return response;
        }
    }

    @Nullable
    private static String getResponseBody(@NotNull HttpResponse response) throws IOException, JsonSyntaxException, IllegalStateException {
        HttpEntity entity = response.getEntity();
        return EntityUtils.toString(entity, StandardCharsets.UTF_8);
    }

    private static int getStatusCode(@NotNull HttpResponse response) {
        return response.getStatusLine().getStatusCode();
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
