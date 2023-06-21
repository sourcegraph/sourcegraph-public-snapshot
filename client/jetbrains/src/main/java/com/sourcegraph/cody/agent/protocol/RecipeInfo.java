package com.sourcegraph.cody.agent.protocol;

public class RecipeInfo {
  public String id;
  public String title;

  @Override
  public String toString() {
    return "RecipeInfo{" + "id='" + id + '\'' + ", title='" + title + '\'' + '}';
  }
}
