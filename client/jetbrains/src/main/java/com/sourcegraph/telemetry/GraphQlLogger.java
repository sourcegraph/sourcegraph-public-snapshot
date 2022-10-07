package com.sourcegraph.telemetry;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.api.GraphQlClient;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.SettingsComponent;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.io.IOException;
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

    // This could be exposed later (as public), but currently, we don't use it externally.
    private static void logEvent(Project project, @NotNull Event event, @Nullable Consumer<Integer> callback) {
        String instanceUrl = ConfigUtil.getSourcegraphUrl(project);
        String accessToken = ConfigUtil.getInstanceType(project) == SettingsComponent.InstanceType.ENTERPRISE ? ConfigUtil.getAccessToken(project) : null;
        String customRequestHeaders = ConfigUtil.getCustomRequestHeaders(project);
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
                int responseStatusCode = GraphQlClient.callGraphQLService(instanceUrl, accessToken, customRequestHeaders, query, variables).getStatusCode();
                if (callback != null) {
                    callback.accept(responseStatusCode);
                }
            } catch (IOException e) {
                logger.info(e);
            }
        }).start();
    }
}
