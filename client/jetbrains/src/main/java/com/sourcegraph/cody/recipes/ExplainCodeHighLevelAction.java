package com.sourcegraph.cody.recipes;

public class ExplainCodeHighLevelAction extends SimpleRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new ExplainCodeHighLevelPromptProvider();
  }

  @Override
  protected String getActionComponentName() {
    return "recipe:explain-code-high-level";
  }
}
