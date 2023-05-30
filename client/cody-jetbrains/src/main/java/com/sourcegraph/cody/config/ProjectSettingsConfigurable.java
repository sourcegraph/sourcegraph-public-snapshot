package com.sourcegraph.cody.config;

import com.intellij.openapi.options.Configurable;
import com.intellij.openapi.project.Project;
import javax.swing.*;
import org.jetbrains.annotations.Nls;
import org.jetbrains.annotations.Nullable;

/** Provides controller functionality for application settings. */
public class ProjectSettingsConfigurable implements Configurable {
  private final Project project;
  private SettingsComponent mySettingsComponent;

  public ProjectSettingsConfigurable(Project project) {
    this.project = project;
  }

  @Nls(capitalization = Nls.Capitalization.Title)
  @Override
  public String getDisplayName() {
    return "Cody (Project)";
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
    return SettingsConfigurableHelper.isModified(project, mySettingsComponent);
  }

  @Override
  public void apply() {
    SettingsConfigurableHelper.apply(project, mySettingsComponent);
  }

  @Override
  public void reset() {
    SettingsConfigurableHelper.reset(project, mySettingsComponent);
  }

  @Override
  public void disposeUIResources() {
    mySettingsComponent.dispose();
    mySettingsComponent = null;
  }
}
