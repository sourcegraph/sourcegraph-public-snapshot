package com.sourcegraph.cody.config;

import com.intellij.openapi.options.Configurable;
import javax.swing.*;
import org.jetbrains.annotations.Nls;
import org.jetbrains.annotations.Nullable;

/** Provides controller functionality for application settings. */
public class ApplicationSettingsConfigurable implements Configurable {
  private SettingsComponent mySettingsComponent;

  @Nls(capitalization = Nls.Capitalization.Title)
  @Override
  public String getDisplayName() {
    return "Cody";
  }

  @Override
  public JComponent getPreferredFocusedComponent() {
    return mySettingsComponent.getPreferredFocusedComponent();
  }

  @Nullable
  @Override
  public JComponent createComponent() {
    mySettingsComponent = new SettingsComponent();
    return mySettingsComponent.getPanel();
  }

  @Override
  public boolean isModified() {
    return SettingsConfigurableHelper.isModified(null, mySettingsComponent);
  }

  @Override
  public void apply() {
    SettingsConfigurableHelper.apply(null, mySettingsComponent);
  }

  @Override
  public void reset() {
    SettingsConfigurableHelper.reset(null, mySettingsComponent);
  }

  @Override
  public void disposeUIResources() {
    mySettingsComponent.dispose();
    mySettingsComponent = null;
  }
}
