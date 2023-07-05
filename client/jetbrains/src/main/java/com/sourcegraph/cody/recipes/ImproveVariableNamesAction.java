package com.sourcegraph.cody.recipes;

public class ImproveVariableNamesAction extends SimpleRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new ImproveVariableNamesPromptProvider();
  }

  @Override
  protected String getActionComponentName() {
    return "recipe:improve-variable-names";
  }
}
