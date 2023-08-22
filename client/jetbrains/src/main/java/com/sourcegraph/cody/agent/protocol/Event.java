package com.sourcegraph.cody.agent.protocol;

import com.google.gson.JsonObject;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class Event {
  @NotNull public String event;
  @NotNull public String userCookieID;
  @NotNull public String url;
  @NotNull public String source;
  @NotNull public String referrer;
  @NotNull public Object argument;
  @Nullable public Object publicArgument;
  @NotNull public String client;
  @Nullable public String connectedSiteID;
  @Nullable public String hashedLicenseKey;
  @NotNull final String deviceID;

  public Event(
      @NotNull String eventName,
      @NotNull String anonymousUserId,
      @NotNull String url,
      @Nullable JsonObject eventProperties,
      @Nullable JsonObject publicArgument) {
    this.event = eventName;
    this.userCookieID = anonymousUserId;
    this.url = url;
    this.source = "IDEEXTENSION";
    this.referrer = "JETBRAINS";
    if (eventProperties != null) {
      this.argument = eventProperties;
    }
    if (publicArgument != null) {
      this.publicArgument = publicArgument;
    }
    this.deviceID = anonymousUserId;
  }
}
