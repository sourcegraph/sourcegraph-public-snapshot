package com.sourcegraph.cody.recipes;

public class GenerateDocStringAction extends SimpleRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new GenerateDocStringPromptProvider();
  }

  @Override
  protected String getActionComponentName() {
    return "recipe:generate-docstring";
  }
}
