package com.sourcegraph.telemetry;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonObject;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.PluginUtil;
import com.sourcegraph.cody.agent.CodyAgent;
import com.sourcegraph.cody.agent.protocol.CompletionEvent;
import com.sourcegraph.cody.agent.protocol.Event;
import com.sourcegraph.config.ConfigUtil;
import java.util.concurrent.CompletableFuture;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class GraphQlLogger {
  private static final Gson gson = new GsonBuilder().serializeNulls().create();

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
      @NotNull Project project,
      long latencyMs,
      long displayDurationMs,
      CompletionEvent.@Nullable Params params) {
    String eventName = "CodyJetBrainsPlugin:completion:suggested";
    JsonObject eventParameters = new JsonObject();
    eventParameters.addProperty("latency", latencyMs);
    eventParameters.addProperty("displayDuration", displayDurationMs);
    eventParameters.addProperty("isAnyKnownPluginEnabled", PluginUtil.isAnyKnownPluginEnabled());
    JsonObject updatedEventParameters = addCompletionEventParams(eventParameters, params);
    logEvent(project, createEvent(project, eventName, updatedEventParameters));
  }

  public static void logAutocompleteAcceptedEvent(
      @NotNull Project project, @Nullable CompletionEvent.Params params) {
    String eventName = "CodyJetBrainsPlugin:completion:accepted";
    JsonObject eventParameters = addCompletionEventParams(new JsonObject(), params);
    logEvent(project, createEvent(project, eventName, eventParameters));
  }

  private static JsonObject addCompletionEventParams(
      JsonObject eventParameters, CompletionEvent.@Nullable Params params) {
    var updatedEventParameters = eventParameters.deepCopy();
    if (params != null) {
      if (params.contextSummary != null) {
        updatedEventParameters.add("contextSummary", gson.toJsonTree(params.contextSummary));
      }
      updatedEventParameters.addProperty("id", params.id);
      updatedEventParameters.addProperty("languageId", params.languageId);
      updatedEventParameters.addProperty("source", params.source);
      updatedEventParameters.addProperty("charCount", params.charCount);
      updatedEventParameters.addProperty("lineCount", params.lineCount);
      updatedEventParameters.addProperty("multilineMode", params.multilineMode);
      updatedEventParameters.addProperty("providerIdentifier", params.providerIdentifier);
    }
    return updatedEventParameters;
  }

  public static void logCodyEvent(
      @NotNull Project project, @NotNull String componentName, @NotNull String action) {
    var eventName = "CodyJetBrainsPlugin:" + componentName + ":" + action;
    logEvent(project, createEvent(project, eventName, new JsonObject()));
  }

  @NotNull
  private static Event createEvent(
      @NotNull Project project, @NotNull String eventName, @NotNull JsonObject eventParameters) {
    var updatedEventParameters = addGlobalEventParameters(eventParameters, project);
    String anonymousUserId = ConfigUtil.getAnonymousUserId();
    return new Event(
        eventName, anonymousUserId != null ? anonymousUserId : "", "", updatedEventParameters);
  }

  @NotNull
  private static JsonObject addGlobalEventParameters(
      @NotNull JsonObject eventParameters, @NotNull Project project) {
    // project specific properties
    var updatedEventParameters = eventParameters.deepCopy();
    updatedEventParameters.addProperty("serverEndpoint", ConfigUtil.getSourcegraphUrl(project));
    // Extension specific properties
    JsonObject extensionDetails = new JsonObject();
    extensionDetails.addProperty("ide", "JetBrains");
    extensionDetails.addProperty("ideExtensionType", "Cody");
    extensionDetails.addProperty("version", ConfigUtil.getPluginVersion());
    updatedEventParameters.add("extensionDetails", extensionDetails);
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
