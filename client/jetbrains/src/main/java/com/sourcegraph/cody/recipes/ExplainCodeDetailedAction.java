package com.sourcegraph.cody.recipes;

public class ExplainCodeDetailedAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new ExplainCodeDetailedPromptProvider();
  }
}
