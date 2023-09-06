package com.sourcegraph.config;

public class PluginSettingChangeContext {
  public final boolean oldCodyEnabled;
  public final boolean oldCodyAutocompleteEnabled;
  public final boolean serverUrlChanged;
  public final boolean newCodyEnabled;
  public final boolean newCodyAutocompleteEnabled;
  public boolean accessTokenChanged;

  public PluginSettingChangeContext(
      boolean serverUrlChanged,
      boolean accessTokenChanged,
      boolean oldCodyEnabled,
      boolean newCodyEnabled,
      boolean oldCodyAutocompleteEnabled,
      boolean newCodyAutocompleteEnabled) {
    this.oldCodyEnabled = oldCodyEnabled;
    this.oldCodyAutocompleteEnabled = oldCodyAutocompleteEnabled;
    this.serverUrlChanged = serverUrlChanged;
    this.newCodyEnabled = newCodyEnabled;
    this.newCodyAutocompleteEnabled = newCodyAutocompleteEnabled;
    this.accessTokenChanged = accessTokenChanged;
  }
}
