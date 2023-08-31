package com.sourcegraph.config;

import org.jetbrains.annotations.NotNull;

public class PluginSettingChangeContext {
  @NotNull public final String oldUrl;
  public final boolean oldCodyEnabled;
  public final boolean oldCodyAutocompleteEnabled;

  @NotNull public final String newUrl;
  public final boolean isAuthMethodChanged;
  public final boolean newCodyEnabled;
  public final boolean newCodyAutocompleteEnabled;

  public PluginSettingChangeContext(
      boolean oldCodyEnabled,
      boolean oldCodyAutocompleteEnabled,
      @NotNull String oldUrl,
      @NotNull String newUrl,
      boolean isAuthMethodChanged,
      boolean newCodyEnabled,
      boolean newCodyAutocompleteEnabled) {
    this.oldCodyEnabled = oldCodyEnabled;
    this.oldCodyAutocompleteEnabled = oldCodyAutocompleteEnabled;
    this.oldUrl = oldUrl;
    this.newUrl = newUrl;
    this.isAuthMethodChanged = isAuthMethodChanged;
    this.newCodyEnabled = newCodyEnabled;
    this.newCodyAutocompleteEnabled = newCodyAutocompleteEnabled;
  }
}
