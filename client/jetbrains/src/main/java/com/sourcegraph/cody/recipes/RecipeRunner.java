package com.sourcegraph.cody.recipes;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.TruncationUtils;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import com.sourcegraph.cody.prompts.LanguageUtils;
import java.util.*;
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
        new TruncatedText(
            TruncationUtils.truncateText(
                textInputToPrompt, TruncationUtils.MAX_RECIPE_INPUT_TOKENS));

    OriginalText selectedText = new OriginalText(textInputToPrompt);
    String truncatedPrecedingText =
        editorContext.getPrecedingText() != null
            ? TruncationUtils.truncateTextStart(
                editorContext.getPrecedingText(), TruncationUtils.MAX_RECIPE_SURROUNDING_TOKENS)
            : "";
    String truncatedFollowingText =
        editorContext.getFollowingText() != null
            ? TruncationUtils.truncateText(
                editorContext.getFollowingText(), TruncationUtils.MAX_RECIPE_SURROUNDING_TOKENS)
            : "";

    PromptContext promptContext =
        promptProvider.getPromptContext(language, selectedText, truncatedTextInputToPrompt);

    ChatMessage humanMessage =
        ChatMessage.createHumanMessage(
            promptContext.getPrompt(), promptContext.getDisplayText(), Collections.emptyList());

    chat.respondToMessage(humanMessage, promptContext.getResponsePrefix());
  }

  public void runGitHistory() {}

  public void runFixup() {}

  public void runContextSearch() {}

  public void runReleaseNotes() {}
}
