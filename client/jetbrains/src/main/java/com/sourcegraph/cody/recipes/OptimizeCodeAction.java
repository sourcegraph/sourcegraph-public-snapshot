package com.sourcegraph.cody.recipes;

public class OptimizeCodeAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new OptimizeCodePromptProvider();
  }
}
