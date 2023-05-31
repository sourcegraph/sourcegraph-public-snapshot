package com.sourcegraph.cody.config;

import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
  @Nullable public final String oldDotcomAccessToken;

  @Nullable public final String oldEnterpriseUrl;

  @Nullable public final String oldEnterpriseAccessToken;

  @Nullable public final String newDotcomAccessToken;

  @Nullable public final String newEnterpriseUrl;

  @Nullable public final String newEnterpriseAccessToken;

  @Nullable public final String newCustomRequestHeaders;

  public PluginSettingChangeContext(
      @Nullable String oldDotcomAccessToken,
      @Nullable String oldEnterpriseUrl,
      @Nullable String oldEnterpriseAccessToken,
      @Nullable String newDotcomAccessToken,
      @Nullable String newEnterpriseUrl,
      @Nullable String newEnterpriseAccessToken,
      @Nullable String newCustomRequestHeaders) {
    this.oldDotcomAccessToken = oldDotcomAccessToken;
    this.oldEnterpriseUrl = oldEnterpriseUrl;
    this.oldEnterpriseAccessToken = oldEnterpriseAccessToken;
    this.newDotcomAccessToken = newDotcomAccessToken;
    this.newEnterpriseUrl = newEnterpriseUrl;
    this.newEnterpriseAccessToken = newEnterpriseAccessToken;
    this.newCustomRequestHeaders = newCustomRequestHeaders;
  }
}
