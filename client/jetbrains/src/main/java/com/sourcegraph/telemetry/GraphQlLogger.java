package com.sourcegraph.telemetry;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.api.GraphQlClient;
import com.sourcegraph.config.ConfigUtil;
import com.sourcegraph.config.SettingsComponent;
import java.io.IOException;
import java.util.Optional;
import java.util.function.Consumer;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class GraphQlLogger {
  private static final Logger logger = Logger.getInstance(GraphQlLogger.class);

  public static void logInstallEvent(
      @NotNull Project project, @NotNull Consumer<Boolean> callback) {
    if (ConfigUtil.getAnonymousUserId() != null) {
      var event = createEvent(project, "CodyInstalled", new JsonObject());
      logEvent(project, event, (responseStatusCode) -> callback.accept(responseStatusCode == 200));
    }
  }

  public static void logUninstallEvent(@NotNull Project project) {
    if (ConfigUtil.getAnonymousUserId() != null) {
      Event event = createEvent(project, "CodyUninstalled", new JsonObject());
      logEvent(project, event, null);
    }
  }

  public static void logAutocompleteSuggestedEvent(
      @NotNull Project project, long latencyMs, long displayDurationMs) {
    String eventName = "CodyJetBrainsPlugin:completion:suggested";
    JsonObject eventParameters = new JsonObject();
    eventParameters.addProperty("latency", latencyMs);
    eventParameters.addProperty("displayDuration", displayDurationMs);
    logEvent(project, createEvent(project, eventName, eventParameters), null);
  }

  public static void logCodyEvent(
      @NotNull Project project, @NotNull String componentName, @NotNull String action) {
    var eventName = "CodyJetBrainsPlugin:" + componentName + ":" + action;
    logEvent(project, createEvent(project, eventName, new JsonObject()), null);
  }

  @NotNull
  private static Event createEvent(
      @NotNull Project project, @NotNull String eventName, @NotNull JsonObject eventParameters) {
    var updatedEventParameters = addProjectSpecificEventParameters(eventParameters, project);
    String anonymousUserId = ConfigUtil.getAnonymousUserId();
    return new Event(
        eventName,
        anonymousUserId != null ? anonymousUserId : "",
        "",
        updatedEventParameters,
        updatedEventParameters);
  }

  @NotNull
  private static JsonObject addProjectSpecificEventParameters(
      @NotNull JsonObject eventParameters, @NotNull Project project) {
    var updatedEventParameters = eventParameters.deepCopy();
    updatedEventParameters.addProperty("serverEndpoint", ConfigUtil.getSourcegraphUrl(project));
    return updatedEventParameters;
  }

  private static void callGraphQlService(
      @Nullable Consumer<Integer> callback,
      @NotNull String instanceUrl,
      @Nullable String accessToken,
      @Nullable String requestHeaders,
      @NotNull String query,
      @NotNull JsonObject variables) {
    new Thread(
            () -> {
              try {
                int responseStatusCode =
                    GraphQlClient.callGraphQLService(
                            instanceUrl, accessToken, requestHeaders, query, variables)
                        .getStatusCode();
                Optional.ofNullable(callback).ifPresent(c -> c.accept(responseStatusCode));
              } catch (IOException e) {
                logger.info(e);
              }
            })
        .start();
  }

  private static void logDotcomEvent(
      @Nullable Consumer<Integer> callback, @NotNull String query, @NotNull JsonObject variables) {
    String instanceUrl = ConfigUtil.DOTCOM_URL;
    callGraphQlService(callback, instanceUrl, null, null, query, variables);
  }

  private static void logInstanceEvent(
      @Nullable Consumer<Integer> callback,
      @NotNull String query,
      @NotNull JsonObject variables,
      @NotNull Project project) {
    String instanceUrl = ConfigUtil.getSourcegraphUrl(project);
    String accessToken = ConfigUtil.getProjectAccessToken(project);
    String customRequestHeaders = ConfigUtil.getCustomRequestHeaders(project);
    callGraphQlService(callback, instanceUrl, accessToken, customRequestHeaders, query, variables);
  }

  // This could be exposed later (as public), but currently, we don't use it externally.
  private static void logEvent(
      @NotNull Project project, @NotNull Event event, @Nullable Consumer<Integer> callback) {
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
    SettingsComponent.InstanceType instanceType = ConfigUtil.getInstanceType(project);
    logDotcomEvent(callback, query, variables); // always log events to dotcom
    if (instanceType != SettingsComponent.InstanceType.DOTCOM // also log to the instance separately
        // but only if its url is actually different from dotcom
        && !ConfigUtil.getSourcegraphUrl(project).equals(ConfigUtil.DOTCOM_URL)) {
      logInstanceEvent(callback, query, variables, project);
    }
  }
}
