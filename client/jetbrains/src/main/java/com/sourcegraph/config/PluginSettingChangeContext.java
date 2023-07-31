package com.sourcegraph.config;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
  @NotNull public final String oldUrl;
  public final boolean oldCodyEnabled;
  public final boolean oldCodyAutoCompleteEnabled;

  @NotNull public final String newUrl;
  public final boolean isDotComAccessTokenChanged;
  public final boolean isEnterpriseAccessTokenChanged;

  @Nullable public final String newCustomRequestHeaders;
  public final boolean newCodyEnabled;
  public final boolean newCodyAutoCompleteEnabled;

  public PluginSettingChangeContext(
      boolean oldCodyEnabled,
      boolean oldCodyAutoCompleteEnabled,
      @NotNull String oldUrl,
      @NotNull String newUrl,
      boolean isDotComAccessTokenChanged,
      boolean isEnterpriseAccessTokenChanged,
      @Nullable String newCustomRequestHeaders,
      boolean newCodyEnabled,
      boolean newCodyAutoCompleteEnabled) {
    this.oldCodyEnabled = oldCodyEnabled;
    this.oldCodyAutoCompleteEnabled = oldCodyAutoCompleteEnabled;
    this.oldUrl = oldUrl;
    this.newUrl = newUrl;
    this.isDotComAccessTokenChanged = isDotComAccessTokenChanged;
    this.isEnterpriseAccessTokenChanged = isEnterpriseAccessTokenChanged;
    this.newCustomRequestHeaders = newCustomRequestHeaders;
    this.newCodyEnabled = newCodyEnabled;
    this.newCodyAutoCompleteEnabled = newCodyAutoCompleteEnabled;
  }
}
