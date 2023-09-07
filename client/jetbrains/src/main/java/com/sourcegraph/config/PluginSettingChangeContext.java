package com.sourcegraph.config;

import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
  public final boolean oldCodyEnabled;
  public final boolean oldCodyAutocompleteEnabled;
  public final boolean serverUrlChanged;
  public final boolean newCodyEnabled;
  public final boolean newCodyAutocompleteEnabled;
  public boolean accessTokenChanged;
  public final boolean oldIsCustomAutocompleteColorEnabled;
  public final boolean isCustomAutocompleteColorEnabled;
  @Nullable public final Integer oldCustomAutocompleteColor;
  @Nullable public final Integer customAutocompleteColor;

  public PluginSettingChangeContext(
      boolean serverUrlChanged,
      boolean accessTokenChanged,
      boolean oldCodyEnabled,
      boolean newCodyEnabled,
      boolean oldCodyAutocompleteEnabled,
      boolean newCodyAutocompleteEnabled,
      @Nullable Integer oldCustomAutocompleteColor,
      boolean oldIsCustomAutocompleteColorEnabled,
      boolean isCustomAutocompleteColorEnabled,
      @Nullable Integer customAutocompleteColor) {
    this.oldCodyEnabled = oldCodyEnabled;
    this.oldCodyAutocompleteEnabled = oldCodyAutocompleteEnabled;
    this.serverUrlChanged = serverUrlChanged;
    this.newCodyEnabled = newCodyEnabled;
    this.newCodyAutocompleteEnabled = newCodyAutocompleteEnabled;
    this.accessTokenChanged = accessTokenChanged;
    this.isCustomAutocompleteColorEnabled = isCustomAutocompleteColorEnabled;
    this.oldIsCustomAutocompleteColorEnabled = oldIsCustomAutocompleteColorEnabled;
    this.oldCustomAutocompleteColor = oldCustomAutocompleteColor;
    this.customAutocompleteColor = customAutocompleteColor;
  }
}
