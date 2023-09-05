package com.sourcegraph.config;

import java.util.Objects;

public class UserAccount {
  private final SettingsComponent.InstanceType instanceType;
  private final String url;
  private final boolean isSelected;

  public UserAccount(SettingsComponent.InstanceType instanceType, String url, boolean isSelected) {
    this.instanceType = instanceType;
    this.url = url;
    this.isSelected = isSelected;
  }

  public SettingsComponent.InstanceType getInstanceType() {
    return instanceType;
  }

  public String getUrl() {
    return url;
  }

  public boolean isSelected() {
    return isSelected;
  }

  public UserAccount asSelected() {
    return new UserAccount(instanceType, url, true);
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;
    UserAccount that = (UserAccount) o;
    return isSelected == that.isSelected
        && instanceType == that.instanceType
        && Objects.equals(url, that.url);
  }

  @Override
  public int hashCode() {
    return Objects.hash(instanceType, url, isSelected);
  }
}
