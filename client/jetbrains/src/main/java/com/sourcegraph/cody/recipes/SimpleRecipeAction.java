package com.sourcegraph.cody.recipes;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.wm.ToolWindow;
import com.intellij.openapi.wm.ToolWindowManager;
import com.sourcegraph.cody.CodyToolWindowContent;
import com.sourcegraph.cody.CodyToolWindowFactory;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.telemetry.GraphQlLogger;
import org.jetbrains.annotations.NotNull;

public abstract class SimpleRecipeAction extends BaseRecipeAction {
  @Override
  public void actionPerformed(@NotNull AnActionEvent e) {
    Project project = e.getProject();
    if (project == null) {
      return;
    }

    ToolWindow toolWindow =
        ToolWindowManager.getInstance(project).getToolWindow(CodyToolWindowFactory.TOOL_WINDOW_ID);
    CodyToolWindowContent codyToolWindowContent = CodyToolWindowContent.getInstance(project);
    if (toolWindow != null) {
      if (!toolWindow.isVisible()) {
        toolWindow.show();
      }
    }
    if (codyToolWindowContent != null) {
      executeRecipeWithPromptProvider(codyToolWindowContent, project);
    }
  }

  public void executeRecipeWithPromptProvider(
      @NotNull UpdatableChat updatableChat, @NotNull Project project) {
    GraphQlLogger.logCodyEvent(project, this.getActionComponentName(), "clicked");
    RecipeRunner recipeRunner = new RecipeRunner(project, updatableChat);
    ActionUtil.runIfCodeSelected(
        updatableChat,
        project,
        (editorSelection) -> recipeRunner.runRecipe(this.getPromptProvider(), editorSelection));
  }

  protected abstract PromptProvider getPromptProvider();

  protected abstract String getActionComponentName();
}
