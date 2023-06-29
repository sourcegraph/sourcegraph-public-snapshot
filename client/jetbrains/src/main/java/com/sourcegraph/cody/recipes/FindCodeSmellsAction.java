package com.sourcegraph.cody.recipes;

/**
 * This action is intentionally not registered in the plugin.xml file. It is used only as a recipe,
 * it's not available in the context menu.
 */
public class FindCodeSmellsAction extends BaseRecipeAction {

  @Override
  protected PromptProvider getPromptProvider() {
    return new FindCodeSmellsPromptProvider();
  }
}
