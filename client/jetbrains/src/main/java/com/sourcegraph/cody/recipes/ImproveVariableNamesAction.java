package com.sourcegraph.cody.recipes;

public class ImproveVariableNamesAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new ImproveVariableNamesPromptProvider();
  }
}
