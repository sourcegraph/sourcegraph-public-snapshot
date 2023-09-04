package com.sourcegraph.config;

import org.jetbrains.annotations.NotNull;

public class PluginSettingChangeContext {
  @NotNull public final String oldUrl;
  public final boolean oldCodyEnabled;
  public final boolean oldCodyAutocompleteEnabled;

  @NotNull public final String newUrl;
  public final boolean newCodyEnabled;
  public final boolean newCodyAutocompleteEnabled;
  public boolean accessTokenChanged;

  public PluginSettingChangeContext(
      boolean oldCodyEnabled,
      boolean oldCodyAutocompleteEnabled,
      @NotNull String oldUrl,
      @NotNull String newUrl,
      boolean newCodyEnabled,
      boolean newCodyAutocompleteEnabled,
      boolean accessTokenChanged) {
    this.oldCodyEnabled = oldCodyEnabled;
    this.oldCodyAutocompleteEnabled = oldCodyAutocompleteEnabled;
    this.oldUrl = oldUrl;
    this.newUrl = newUrl;
    this.newCodyEnabled = newCodyEnabled;
    this.newCodyAutocompleteEnabled = newCodyAutocompleteEnabled;
    this.accessTokenChanged = accessTokenChanged;
  }
}
