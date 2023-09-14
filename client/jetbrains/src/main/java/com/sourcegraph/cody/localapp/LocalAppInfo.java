package com.sourcegraph.cody.localapp;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import org.jetbrains.annotations.Nullable;

/** This is just a data object for the local Cody app.json */
@JsonIgnoreProperties(
    ignoreUnknown = true) // we don't care if the json contains fields we don't know about
@JsonInclude(JsonInclude.Include.NON_NULL)
public class LocalAppInfo {
  @JsonProperty("token")
  @Nullable
  private String token;

  @JsonProperty("endpoint")
  @Nullable
  private String endpoint;

  @JsonProperty("version")
  @Nullable
  private String version;

  @Nullable
  public String getToken() {
    return token;
  }

  public void setToken(@Nullable String token) {
    this.token = token;
  }

  @Nullable
  public String getEndpoint() {
    return endpoint;
  }

  public void setEndpoint(@Nullable String endpoint) {
    this.endpoint = endpoint;
  }

  public @Nullable String getVersion() {
    return version;
  }

  public void setVersion(@Nullable String version) {
    this.version = version;
  }
}
