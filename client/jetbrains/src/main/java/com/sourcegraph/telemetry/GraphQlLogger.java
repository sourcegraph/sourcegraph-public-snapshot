package com.sourcegraph.telemetry;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.config.ConfigUtil;
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
import java.util.function.Consumer;

public class GraphQlLogger {
    private static final Logger logger = Logger.getInstance(GraphQlLogger.class);

    public static void logInstallEvent(Project project, Consumer<Boolean> callback) {
        String anonymousUserId = ConfigUtil.getAnonymousUserId();
        if (anonymousUserId != null) {
            Event event = new Event("IDEInstalled", anonymousUserId, ConfigUtil.getSourcegraphUrl(project), null, null);
            logEvent(project, event, (responseStatusCode) -> callback.accept(responseStatusCode == 200));
        }
    }

    public static void logUninstallEvent(Project project) {
        String anonymousUserId = ConfigUtil.getAnonymousUserId();
        if (anonymousUserId != null) {
            Event event = new Event("IDEUninstalled", anonymousUserId, ConfigUtil.getSourcegraphUrl(project), null, null);
            logEvent(project, event, null);
        }
    }

    // This could be exposed later as public, but currently, we don't use it externally.
    private static void logEvent(Project project, @NotNull Event event, @Nullable Consumer<Integer> callback) {
        String instanceUrl = ConfigUtil.getSourcegraphUrl(project);
        String accessToken = ConfigUtil.getAccessToken(project);
        new Thread(() -> {
            String query = "" +
                "mutation LogEvents($events: [Event!]) {" +
                "    logEvents(events: $events) { " +
                "        alwaysNil" +
                "    }" +
                "}";

            JsonArray events = new JsonArray();
            events.add(event.toJson());
            JsonObject variables = new JsonObject();
            variables.add("events", events);

            try {
                int responseStatusCode = callGraphQLService(instanceUrl, accessToken, query, variables);
                if (callback != null) {
                    callback.accept(responseStatusCode);
                }
            } catch (IOException e) {
                logger.info(e);
            }
        }).start();
    }

    private static int callGraphQLService(@NotNull String instanceUrl, @Nullable String accessToken, @NotNull String query, @NotNull JsonObject variables) throws IOException {
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
