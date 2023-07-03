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

  public static void logInstallEvent(Project project, Consumer<Boolean> callback) {
    String anonymousUserId = ConfigUtil.getAnonymousUserId();
    JsonObject eventParameters = getEventParameters(project);
    if (anonymousUserId != null) {
      Event event =
          new Event("CodyInstalled", anonymousUserId, "", eventParameters, eventParameters);
      logEvent(project, event, (responseStatusCode) -> callback.accept(responseStatusCode == 200));
    }
  }

  public static void logUninstallEvent(Project project) {
    String anonymousUserId = ConfigUtil.getAnonymousUserId();
    JsonObject eventParameters = getEventParameters(project);
    if (anonymousUserId != null) {
      Event event =
          new Event("CodyUninstalled", anonymousUserId, "", eventParameters, eventParameters);
      logEvent(project, event, null);
    }
  }

  public static void logCodyEvents(
      @NotNull Project project, @NotNull String componentName, String... actions) {
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
