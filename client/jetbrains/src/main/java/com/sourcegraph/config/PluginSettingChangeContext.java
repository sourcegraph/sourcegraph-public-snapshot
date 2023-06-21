package com.sourcegraph.config;

import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
  @Nullable public final String oldUrl;

  @Nullable public final String oldDotComAccessToken;
  @Nullable public final String oldEnterpriseAccessToken;
  public final boolean oldCodyCompletionsEnabled;

  @Nullable public final String newUrl;

  @Nullable public final String newDotComAccessToken;
  @Nullable public final String newEnterpriseAccessToken;

  @Nullable public final String newCustomRequestHeaders;
  public final boolean newCodyCompletionsEnabled;

  public PluginSettingChangeContext(
      @Nullable String oldUrl,
      @Nullable String oldDotComAccessToken,
      @Nullable String oldEnterpriseAccessToken,
      boolean oldCodyCompletionsEnabled,
      @Nullable String newUrl,
      @Nullable String newDotComAccessToken,
      @Nullable String newEnterpriseAccessToken,
      @Nullable String newCustomRequestHeaders,
      boolean newCodyCompletionsEnabled) {
    this.oldUrl = oldUrl;
    this.oldDotComAccessToken = oldDotComAccessToken;
    this.oldEnterpriseAccessToken = oldEnterpriseAccessToken;
    this.oldCodyCompletionsEnabled = oldCodyCompletionsEnabled;
    this.newUrl = newUrl;
    this.newDotComAccessToken = newDotComAccessToken;
    this.newEnterpriseAccessToken = newEnterpriseAccessToken;
    this.newCustomRequestHeaders = newCustomRequestHeaders;
    this.newCodyCompletionsEnabled = newCodyCompletionsEnabled;
  }
}
