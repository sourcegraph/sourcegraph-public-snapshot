package com.sourcegraph.cody.recipes;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.UpdatableChatHolderService;
import org.jetbrains.annotations.NotNull;

public abstract class BaseRecipeAction extends DumbAwareAction {
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
    RecipeRunner recipeRunner = new RecipeRunner(project, updatableChat);
    ActionUtil.runIfCodeSelected(
        updatableChat,
        project,
        (editorSelection) -> recipeRunner.runRecipe(this.getPromptProvider(), editorSelection));
  }

  protected abstract PromptProvider getPromptProvider();
}
