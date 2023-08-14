package com.sourcegraph.cody.agent;

import java.util.Map;

public class ConnectionConfiguration {

  public String serverEndpoint;
  public String accessToken;
  public Map<String, String> customHeaders;
  public String autocompleteAdvancedProvider;
  public String autocompleteAdvancedServerEndpoint;
  public String autocompleteAdvancedAccessToken;
  public boolean autocompleteAdvancedEmbeddings;
  public boolean debug;
  public boolean verboseDebug;

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

  public ConnectionConfiguration setAutocompleteAdvancedProvider(
      String autocompleteAdvancedProvider) {
    this.autocompleteAdvancedProvider = autocompleteAdvancedProvider;
    return this;
  }

  public ConnectionConfiguration setAutocompleteAdvancedServerEndpoint(
      String autocompleteAdvancedServerEndpoint) {
    this.autocompleteAdvancedServerEndpoint = autocompleteAdvancedServerEndpoint;
    return this;
  }

  public ConnectionConfiguration setAutocompleteAdvancedAccessToken(
      String autocompleteAdvancedAccessToken) {
    this.autocompleteAdvancedAccessToken = autocompleteAdvancedAccessToken;
    return this;
  }

  public ConnectionConfiguration setAutocompleteAdvancedEmbeddings(
      boolean autocompleteAdvancedEmbeddings) {
    this.autocompleteAdvancedEmbeddings = autocompleteAdvancedEmbeddings;
    return this;
  }

  public ConnectionConfiguration setDebug(boolean debug) {
    this.debug = debug;
    return this;
  }

  public ConnectionConfiguration setVerboseDebug(boolean verboseDebug) {
    this.verboseDebug = verboseDebug;
    return this;
  }

  @Override
  public String toString() {
    return "ConnectionConfiguration{"
        + "serverEndpoint='"
        + serverEndpoint
        + '\''
        + ", accessToken='"
        + accessToken
        + '\''
        + ", customHeaders="
        + customHeaders
        + ", autocompleteAdvancedProvider='"
        + autocompleteAdvancedProvider
        + '\''
        + ", autocompleteAdvancedServerEndpoint='"
        + autocompleteAdvancedServerEndpoint
        + '\''
        + ", autocompleteAdvancedAccessToken='"
        + autocompleteAdvancedAccessToken
        + '\''
        + ", autocompleteAdvancedEmbeddings="
        + autocompleteAdvancedEmbeddings
        + '\''
        + ", debug="
        + debug
        + '\''
        + ", verboseDebug="
        + verboseDebug
        + '}';
  }
}
