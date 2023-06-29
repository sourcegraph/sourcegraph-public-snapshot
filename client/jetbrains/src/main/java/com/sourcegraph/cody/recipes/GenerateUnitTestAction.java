package com.sourcegraph.cody.recipes;

public class GenerateUnitTestAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new GenerateUnitTestPromptProvider();
  }
}
