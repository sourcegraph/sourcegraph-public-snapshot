package com.sourcegraph.cody.recipes;

public class GenerateDocStringAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new GenerateDocStringPromptProvider();
  }
}
