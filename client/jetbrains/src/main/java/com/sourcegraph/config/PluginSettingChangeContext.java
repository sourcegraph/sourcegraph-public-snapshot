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
      boolean oldIsCustomAutocompleteColorEnabled,
      boolean isCustomAutocompleteColorEnabled,
      @Nullable Integer oldCustomAutocompleteColor,
      @Nullable Integer customAutocompleteColor) {
    this.serverUrlChanged = serverUrlChanged;
    this.accessTokenChanged = accessTokenChanged;
    this.oldCodyEnabled = oldCodyEnabled;
    this.newCodyEnabled = newCodyEnabled;
    this.oldCodyAutocompleteEnabled = oldCodyAutocompleteEnabled;
    this.newCodyAutocompleteEnabled = newCodyAutocompleteEnabled;
    this.oldIsCustomAutocompleteColorEnabled = oldIsCustomAutocompleteColorEnabled;
    this.isCustomAutocompleteColorEnabled = isCustomAutocompleteColorEnabled;
    this.oldCustomAutocompleteColor = oldCustomAutocompleteColor;
    this.customAutocompleteColor = customAutocompleteColor;
  }
}
