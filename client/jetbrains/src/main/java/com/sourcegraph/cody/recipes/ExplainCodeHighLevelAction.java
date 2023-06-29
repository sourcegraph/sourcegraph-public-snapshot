package com.sourcegraph.cody.recipes;

public class ExplainCodeHighLevelAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new ExplainCodeHighLevelPromptProvider();
  }
}
