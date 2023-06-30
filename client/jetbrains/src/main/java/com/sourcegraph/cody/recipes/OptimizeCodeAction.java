package com.sourcegraph.cody.recipes;

public class OptimizeCodeAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new OptimizeCodePromptProvider();
  }

  @Override
  protected String getActionComponentName() {
    return "recipe:optimize-code";
  }
}
