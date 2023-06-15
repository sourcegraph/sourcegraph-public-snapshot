package com.sourcegraph.agent;

import java.util.Map;

public class Configuration {

  public String serverEndpoint;
  public String accessToken;
  public Map<String, String> customHeaders;

  public Configuration setServerEndpoint(String serverEndpoint) {
    this.serverEndpoint = serverEndpoint;
    return this;
  }

  public Configuration setAccessToken(String accessToken) {
    this.accessToken = accessToken;
    return this;
  }

  public Configuration setCustomHeaders(Map<String, String> customHeaders) {
    this.customHeaders = customHeaders;
    return this;
  }
}
