package com.sourcegraph.cody.agent;

import java.util.Map;

public class ExtensionConfiguration {

  public String serverEndpoint;
  public String proxy;
  public String accessToken;
  public Map<String, String> customHeaders;
  public String autocompleteAdvancedProvider;
  public String autocompleteAdvancedServerEndpoint;
  public String autocompleteAdvancedAccessToken;
  public boolean autocompleteAdvancedEmbeddings;
  public boolean debug;
  public boolean verboseDebug;
  public String codebase;

  public ExtensionConfiguration setServerEndpoint(String serverEndpoint) {
    this.serverEndpoint = serverEndpoint;
    return this;
  }

  public ExtensionConfiguration setProxy(String proxy) {
    this.proxy = proxy;
    return this;
  }

  public ExtensionConfiguration setAccessToken(String accessToken) {
    this.accessToken = accessToken;
    return this;
  }

  public ExtensionConfiguration setCustomHeaders(Map<String, String> customHeaders) {
    this.customHeaders = customHeaders;
    return this;
  }

  public ExtensionConfiguration setAutocompleteAdvancedProvider(
      String autocompleteAdvancedProvider) {
    this.autocompleteAdvancedProvider = autocompleteAdvancedProvider;
    return this;
  }

  public ExtensionConfiguration setAutocompleteAdvancedServerEndpoint(
      String autocompleteAdvancedServerEndpoint) {
    this.autocompleteAdvancedServerEndpoint = autocompleteAdvancedServerEndpoint;
    return this;
  }

  public ExtensionConfiguration setAutocompleteAdvancedAccessToken(
      String autocompleteAdvancedAccessToken) {
    this.autocompleteAdvancedAccessToken = autocompleteAdvancedAccessToken;
    return this;
  }

  public ExtensionConfiguration setAutocompleteAdvancedEmbeddings(
      boolean autocompleteAdvancedEmbeddings) {
    this.autocompleteAdvancedEmbeddings = autocompleteAdvancedEmbeddings;
    return this;
  }

  public ExtensionConfiguration setDebug(boolean debug) {
    this.debug = debug;
    return this;
  }

  public ExtensionConfiguration setVerboseDebug(boolean verboseDebug) {
    this.verboseDebug = verboseDebug;
    return this;
  }

  public ExtensionConfiguration setCodebase(String codebase) {
    this.codebase = codebase;
    return this;
  }

  @Override
  public String toString() {
    return "ExtensionConfiguration{"
        + "serverEndpoint='"
        + serverEndpoint
        + '\''
        + ", proxy='"
        + proxy
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
        + ", codebase="
        + codebase
        + '}';
  }
}
