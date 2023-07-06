package com.sourcegraph.cody.recipes;

public class GenerateUnitTestAction extends SimpleRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new GenerateUnitTestPromptProvider();
  }

  @Override
  protected String getActionComponentName() {
    return "recipe:generate-unit-test";
  }
}
