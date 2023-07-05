package com.sourcegraph.cody.recipes;

public class ExplainCodeDetailedAction extends SimpleRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new ExplainCodeDetailedPromptProvider();
  }

  @Override
  protected String getActionComponentName() {
    return "recipe:explain-code-detailed";
  }
}
