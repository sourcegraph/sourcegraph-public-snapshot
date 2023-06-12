package com.sourcegraph.config;

import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
  @Nullable public final String oldUrl;

  @Nullable public final String oldDotComAccessToken;
  @Nullable public final String oldEnterpriseAccessToken;

  @Nullable public final String newUrl;

  @Nullable public final String newDotComAccessToken;
  @Nullable public final String newEnterpriseAccessToken;

  @Nullable public final String newCustomRequestHeaders;

  public PluginSettingChangeContext(
      @Nullable String oldUrl,
      @Nullable String oldDotComAccessToken,
      @Nullable String oldEnterpriseAccessToken,
      @Nullable String newUrl,
      @Nullable String newDotComAccessToken,
      @Nullable String newEnterpriseAccessToken,
      @Nullable String newCustomRequestHeaders) {
    this.oldUrl = oldUrl;
    this.oldDotComAccessToken = oldDotComAccessToken;
    this.oldEnterpriseAccessToken = oldEnterpriseAccessToken;
    this.newUrl = newUrl;
    this.newDotComAccessToken = newDotComAccessToken;
    this.newEnterpriseAccessToken = newEnterpriseAccessToken;
    this.newCustomRequestHeaders = newCustomRequestHeaders;
  }
}
