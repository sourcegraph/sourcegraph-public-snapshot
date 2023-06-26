package com.sourcegraph.cody.recipes;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.TruncationUtils;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import com.sourcegraph.cody.prompts.LanguageUtils;
import org.jetbrains.annotations.NotNull;

public class RecipeRunner {
  private final @NotNull Project project;
  private final @NotNull UpdatableChat chat;

  public RecipeRunner(@NotNull Project project, @NotNull UpdatableChat chat) {

    this.project = project;
    this.chat = chat;
  }

  public void runRecipe(@NotNull PromptProvider promptProvider, @NotNull String textInputToPrompt) {
    EditorContext editorContext = EditorContextGetter.getEditorContext(project);
    Language language =
        new Language(
            LanguageUtils.getNormalizedLanguageName(editorContext.getCurrentFileExtension()));

    TruncatedText truncatedTextInputToPrompt =
        TruncatedText.of(textInputToPrompt, TruncationUtils.MAX_RECIPE_INPUT_TOKENS);

    // TODO: Use or remove these
    String precedingText = editorContext.getPrecedingText();
    String truncatedPrecedingText =
        precedingText != null
            ? TruncatedText.ofEndOf(precedingText, TruncationUtils.MAX_RECIPE_SURROUNDING_TOKENS)
                .getValue()
            : "";
    String followingText = editorContext.getFollowingText();
    String truncatedFollowingText =
        followingText != null
            ? TruncatedText.of(followingText, TruncationUtils.MAX_RECIPE_SURROUNDING_TOKENS)
                .getValue()
            : "";

    OriginalText selectedText = new OriginalText(textInputToPrompt);
    PromptContext promptContext =
        promptProvider.getPromptContext(language, selectedText, truncatedTextInputToPrompt);

    ChatMessage humanMessage =
        ChatMessage.createHumanMessage(promptContext.getPrompt(), promptContext.getDisplayText());

    chat.respondToMessage(humanMessage, promptContext.getResponsePrefix());
  }

  // TODO: Implement or remove it
  public void runGitHistory() {}

  public void runFixup() {}

  public void runContextSearch() {}

  public void runReleaseNotes() {}
}
