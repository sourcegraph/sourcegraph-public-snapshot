package com.sourcegraph.cody.agent;

import java.util.Map;

public class ConnectionConfiguration {

  public String serverEndpoint;
  public String accessToken;
  public Map<String, String> customHeaders;

  public ConnectionConfiguration setServerEndpoint(String serverEndpoint) {
    this.serverEndpoint = serverEndpoint;
    return this;
  }

  public ConnectionConfiguration setAccessToken(String accessToken) {
    this.accessToken = accessToken;
    return this;
  }

  public ConnectionConfiguration setCustomHeaders(Map<String, String> customHeaders) {
    this.customHeaders = customHeaders;
    return this;
  }
}
