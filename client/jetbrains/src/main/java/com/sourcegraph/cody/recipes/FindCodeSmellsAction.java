package com.sourcegraph.cody.recipes;

public class FindCodeSmellsAction extends SimpleRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new FindCodeSmellsPromptProvider();
  }

  @Override
  protected String getActionComponentName() {
    return "recipe:find-code-smells";
  }
}
