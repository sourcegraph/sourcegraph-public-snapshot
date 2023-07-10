package com.sourcegraph.cody.recipes;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.UpdatableChatHolderService;
import com.sourcegraph.cody.prompts.SupportedLanguages;
import com.sourcegraph.cody.ui.SelectOptionManager;
import com.sourcegraph.telemetry.GraphQlLogger;
import org.jetbrains.annotations.NotNull;

public class TranslateToLanguageAction extends BaseRecipeAction {
  @Override
  public void actionPerformed(@NotNull AnActionEvent e) {

    Project project = e.getProject();
    if (project == null) {
      return;
    }
    executeAction(project);
  }

  public void executeAction(Project project) {
    GraphQlLogger.logCodyEvent(project, "recipe:translate-to-language", "clicked");
    UpdatableChatHolderService updatableChatHolderService =
        project.getService(UpdatableChatHolderService.class);
    UpdatableChat updatableChat = updatableChatHolderService.getUpdatableChat();
    RecipeRunner recipeRunner = new RecipeRunner(project, updatableChat);
    ActionUtil.runIfCodeSelected(
        updatableChat,
        project,
        (editorSelection) -> {
          SelectOptionManager selectOptionManager = SelectOptionManager.getInstance(project);
          selectOptionManager.show(
              project,
              SupportedLanguages.LANGUAGE_NAMES,
              (selectedLanguage) -> {
                GraphQlLogger.logCodyEvent(project, "recipe:translate-to-language", "executed");
                recipeRunner.runRecipe(
                    new TranslateToLanguagePromptProvider(new Language(selectedLanguage)),
                    editorSelection);
              });
        });
  }
}
