package com.sourcegraph.cody.recipes;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.UpdatableChatHolderService;
import com.sourcegraph.telemetry.GraphQlLogger;
import org.jetbrains.annotations.NotNull;

public abstract class SimpleRecipeAction extends BaseRecipeAction {
  @Override
  public void actionPerformed(@NotNull AnActionEvent e) {
    Project project = e.getProject();
    if (project == null) {
      return;
    }
    UpdatableChatHolderService updatableChatHolderService =
        project.getService(UpdatableChatHolderService.class);
    UpdatableChat updatableChat = updatableChatHolderService.getUpdatableChat();
    executeRecipeWithPromptProvider(updatableChat, project);
  }

  public void executeRecipeWithPromptProvider(UpdatableChat updatableChat, Project project) {
    GraphQlLogger.logCodyEvents(project, this.getActionComponentName(), "clicked");
    RecipeRunner recipeRunner = new RecipeRunner(project, updatableChat);
    ActionUtil.runIfCodeSelected(
        updatableChat,
        project,
        (editorSelection) -> recipeRunner.runRecipe(this.getPromptProvider(), editorSelection));
  }

  protected abstract PromptProvider getPromptProvider();

  protected abstract String getActionComponentName();
}
