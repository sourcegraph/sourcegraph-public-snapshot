package com.sourcegraph.config;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
  @NotNull public final String oldUrl;
  public final boolean oldCodyEnabled;
  public final boolean oldCodyAutoCompleteEnabled;
  public final boolean oldCodyDebugEnabled;
  public final boolean oldCodyVerboseDebugEnabled;

  @NotNull public final String newUrl;
  public final boolean isDotComAccessTokenChanged;
  public final boolean isEnterpriseAccessTokenChanged;

  @Nullable public final String newCustomRequestHeaders;
  public final boolean newCodyEnabled;
  public final boolean newCodyAutoCompleteEnabled;
  public final boolean newCodyDebugEnabled;
  public final boolean newCodyVerboseDebugEnabled;

  public PluginSettingChangeContext(
      boolean oldCodyEnabled,
      boolean oldCodyAutoCompleteEnabled,
      @NotNull String oldUrl,
      boolean oldCodyDebugEnabled,
      boolean oldCodyVerboseDebugEnabled,
      @NotNull String newUrl,
      boolean isDotComAccessTokenChanged,
      boolean isEnterpriseAccessTokenChanged,
      @Nullable String newCustomRequestHeaders,
      boolean newCodyEnabled,
      boolean newCodyAutoCompleteEnabled,
      boolean newCodyDebugEnabled,
      boolean newCodyVerboseDebugEnabled) {
    this.oldCodyEnabled = oldCodyEnabled;
    this.oldCodyAutoCompleteEnabled = oldCodyAutoCompleteEnabled;
    this.oldUrl = oldUrl;
    this.oldCodyDebugEnabled = oldCodyDebugEnabled;
    this.oldCodyVerboseDebugEnabled = oldCodyVerboseDebugEnabled;
    this.newUrl = newUrl;
    this.isDotComAccessTokenChanged = isDotComAccessTokenChanged;
    this.isEnterpriseAccessTokenChanged = isEnterpriseAccessTokenChanged;
    this.newCustomRequestHeaders = newCustomRequestHeaders;
    this.newCodyEnabled = newCodyEnabled;
    this.newCodyAutoCompleteEnabled = newCodyAutoCompleteEnabled;
    this.newCodyDebugEnabled = newCodyDebugEnabled;
    this.newCodyVerboseDebugEnabled = newCodyVerboseDebugEnabled;
  }
}
