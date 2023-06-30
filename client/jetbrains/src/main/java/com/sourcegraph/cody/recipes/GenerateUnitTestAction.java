package com.sourcegraph.cody.recipes;

public class GenerateUnitTestAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new GenerateUnitTestPromptProvider();
  }

  @Override
  protected String getActionComponentName() {
    return "recipe:generate-unit-test";
  }
}
