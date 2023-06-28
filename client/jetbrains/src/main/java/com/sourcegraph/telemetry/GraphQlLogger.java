package com.sourcegraph.telemetry;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.api.GraphQlClient;
import com.sourcegraph.config.ConfigUtil;
import java.io.IOException;
import java.util.function.Consumer;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class GraphQlLogger {
  private static final Logger logger = Logger.getInstance(GraphQlLogger.class);

  public static void logInstallEvent(Project project, Consumer<Boolean> callback) {
    String anonymousUserId = ConfigUtil.getAnonymousUserId();
    JsonObject eventParameters = getEventParameters(project);
    if (anonymousUserId != null) {
      Event event =
          new Event("CodyJetBrainsPlugin:CodyInstalled", anonymousUserId, "", eventParameters, eventParameters);
      logEvent(project, event, (responseStatusCode) -> callback.accept(responseStatusCode == 200));
    }
  }

  public static void logUninstallEvent(Project project) {
    String anonymousUserId = ConfigUtil.getAnonymousUserId();
    JsonObject eventParameters = getEventParameters(project);
    if (anonymousUserId != null) {
      Event event =
          new Event("CodyJetBrainsPlugin:CodyUninstalled", anonymousUserId, "", eventParameters, eventParameters);
      logEvent(project, event, null);
    }
  }

  public static void logCodyEvents(
      @NotNull Project project, @NotNull String componentName, @NotNull String[] actions) {
    for (String action : actions) {
      logCodyEvent(project, componentName, action);
    }
  }

  public static void logCodyEvent(
      @NotNull Project project, @NotNull String componentName, @NotNull String action) {
    String anonymousUserId = ConfigUtil.getAnonymousUserId();
    String eventName = "CodyJetBrainsPlugin:" + componentName + ":" + action;
    JsonObject eventParameters = getEventParameters(project);
    Event event =
        new Event(
            eventName,
            anonymousUserId != null ? anonymousUserId : "",
            "",
            eventParameters,
            eventParameters);
    logEvent(project, event, null);
  }

  @NotNull
  private static JsonObject getEventParameters(@NotNull Project project) {
    JsonObject eventParameters = new JsonObject();
    eventParameters.addProperty("serverEndpoint", ConfigUtil.getSourcegraphUrl(project));
    return eventParameters;
  }

  // This could be exposed later (as public), but currently, we don't use it externally.
  private static void logEvent(
      @NotNull Project project, @NotNull Event event, @Nullable Consumer<Integer> callback) {
    String instanceUrl = ConfigUtil.getSourcegraphUrl(project);
    String accessToken = ConfigUtil.getProjectAccessToken(project);
    String customRequestHeaders = ConfigUtil.getCustomRequestHeaders(project);
    new Thread(
            () -> {
              String query =
                  "mutation LogEvents($events: [Event!]) {"
                      + "    logEvents(events: $events) { "
                      + "        alwaysNil"
                      + "    }"
                      + "}";

              JsonArray events = new JsonArray();
              events.add(event.toJson());
              JsonObject variables = new JsonObject();
              variables.add("events", events);

              try {
                int responseStatusCode =
                    GraphQlClient.callGraphQLService(
                            instanceUrl, accessToken, customRequestHeaders, query, variables)
                        .getStatusCode();
                if (callback != null) {
                  callback.accept(responseStatusCode);
                }
              } catch (IOException e) {
                logger.info(e);
              }
            })
        .start();
  }
}
