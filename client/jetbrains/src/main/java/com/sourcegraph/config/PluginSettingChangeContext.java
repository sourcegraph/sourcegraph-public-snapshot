package com.sourcegraph.config;

import java.awt.*;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
  @NotNull public final String oldUrl;
  public final boolean oldCodyEnabled;
  public final boolean oldCodyAutocompleteEnabled;
  public final boolean oldCodyDebugEnabled;
  public final boolean oldCodyVerboseDebugEnabled;

  @NotNull public final String newUrl;
  public final boolean isDotComAccessTokenChanged;
  public final boolean isEnterpriseAccessTokenChanged;

  @Nullable public final String newCustomRequestHeaders;
  public final boolean newCodyEnabled;
  public final boolean newCodyAutocompleteEnabled;
  public final boolean newCodyDebugEnabled;
  public final boolean newCodyVerboseDebugEnabled;
  public final boolean isCustomAutocompleteColorEnabled;
  public final Color customAutocompleteColor;

  public PluginSettingChangeContext(
      boolean oldCodyEnabled,
      boolean oldCodyAutocompleteEnabled,
      @NotNull String oldUrl,
      boolean oldCodyDebugEnabled,
      boolean oldCodyVerboseDebugEnabled,
      @NotNull String newUrl,
      boolean isDotComAccessTokenChanged,
      boolean isEnterpriseAccessTokenChanged,
      @Nullable String newCustomRequestHeaders,
      boolean newCodyEnabled,
      boolean newCodyAutocompleteEnabled,
      boolean newCodyDebugEnabled,
      boolean newCodyVerboseDebugEnabled,
      boolean isCustomAutocompleteColorEnabled,
      @Nullable Color customAutocompleteColor) {
    this.oldCodyEnabled = oldCodyEnabled;
    this.oldCodyAutocompleteEnabled = oldCodyAutocompleteEnabled;
    this.oldUrl = oldUrl;
    this.oldCodyDebugEnabled = oldCodyDebugEnabled;
    this.oldCodyVerboseDebugEnabled = oldCodyVerboseDebugEnabled;
    this.newUrl = newUrl;
    this.isDotComAccessTokenChanged = isDotComAccessTokenChanged;
    this.isEnterpriseAccessTokenChanged = isEnterpriseAccessTokenChanged;
    this.newCustomRequestHeaders = newCustomRequestHeaders;
    this.newCodyEnabled = newCodyEnabled;
    this.newCodyAutocompleteEnabled = newCodyAutocompleteEnabled;
    this.newCodyDebugEnabled = newCodyDebugEnabled;
    this.newCodyVerboseDebugEnabled = newCodyVerboseDebugEnabled;
    this.isCustomAutocompleteColorEnabled = isCustomAutocompleteColorEnabled;
    this.customAutocompleteColor = customAutocompleteColor;
  }
}
