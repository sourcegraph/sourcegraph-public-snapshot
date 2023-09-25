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

  /**
   * @deprecated This property is no longer used for events.
   *     <p>Use {@link #publicArgument} instead.
   */
  @NotNull public Object argument;

  @Nullable public Object publicArgument;
  @NotNull public final String client;
  @Nullable public String connectedSiteID;
  @Nullable public String hashedLicenseKey;
  @NotNull final String deviceID;

  public Event(
      @NotNull String eventName,
      @NotNull String anonymousUserId,
      @NotNull String url,
      @Nullable JsonObject publicArgument) {
    this.event = eventName;
    this.userCookieID = anonymousUserId;
    this.url = url;
    this.source = "IDEEXTENSION";
    this.referrer = "JETBRAINS";
    this.client = "JETBRAINS_CODY_EXTENSION";
    this.argument = new JsonObject();
    if (publicArgument != null) {
      this.publicArgument = publicArgument;
    }
    this.deviceID = anonymousUserId;
  }
}
