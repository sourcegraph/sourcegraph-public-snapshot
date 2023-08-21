package com.sourcegraph.telemetry;

import com.google.gson.JsonObject;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.PluginUtil;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.protocol.Event;
import com.sourcegraph.config.ConfigUtil;
import java.util.concurrent.CompletableFuture;
import org.jetbrains.annotations.NotNull;

public class GraphQlLogger {

  public static CompletableFuture<Boolean> logInstallEvent(@NotNull Project project) {
    if (ConfigUtil.getAnonymousUserId() != null && project.isDisposed()) {
      var event = createEvent(project, "CodyInstalled", new JsonObject());
      return logEvent(project, event);
    }
    return CompletableFuture.completedFuture(false);
  }

  public static void logUninstallEvent(@NotNull Project project) {
    if (ConfigUtil.getAnonymousUserId() != null) {
      Event event = createEvent(project, "CodyUninstalled", new JsonObject());
      logEvent(project, event);
    }
  }

  public static void logAutocompleteSuggestedEvent(
      @NotNull Project project, long latencyMs, long displayDurationMs) {
    String eventName = "CodyJetBrainsPlugin:completion:suggested";
    JsonObject eventParameters = new JsonObject();
    eventParameters.addProperty("latency", latencyMs);
    eventParameters.addProperty("displayDuration", displayDurationMs);
    eventParameters.addProperty("isAnyKnownPluginEnabled", PluginUtil.isAnyKnownPluginEnabled());
    logEvent(project, createEvent(project, eventName, eventParameters));
  }

  public static void logCodyEvent(
      @NotNull Project project, @NotNull String componentName, @NotNull String action) {
    var eventName = "CodyJetBrainsPlugin:" + componentName + ":" + action;
    logEvent(project, createEvent(project, eventName, new JsonObject()));
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

  // This could be exposed later (as public), but currently, we don't use it externally.
  private static CompletableFuture<Boolean> logEvent(
      @NotNull Project project, @NotNull Event event) {
    return CodyAgent.withServer(project, server -> server.logEvent(event))
        .thenApply((ignored) -> true)
        .exceptionally((ignored) -> false);
  }
}
