package com.sourcegraph.cody.agent.protocol;

import org.jetbrains.annotations.Nullable;

public class RecipeInfo {
  @Nullable public String id;
  @Nullable public String title;

  @Override
  public String toString() {
    return "RecipeInfo{" + "id='" + id + '\'' + ", title='" + title + '\'' + '}';
  }
}
