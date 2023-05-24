package com.sourcegraph.cody.recipes;

import com.intellij.openapi.project.Project;
import com.sourcegraph.cody.TruncationUtils;
import com.sourcegraph.cody.UpdatableChat;
import com.sourcegraph.cody.chat.ChatMessage;
import com.sourcegraph.cody.editor.EditorContext;
import com.sourcegraph.cody.editor.EditorContextGetter;
import com.sourcegraph.cody.prompts.LanguageUtils;
import org.jetbrains.annotations.NotNull;

import java.util.ArrayList;

public class RecipeRunner {
    private final @NotNull Project project;
    private final @NotNull UpdatableChat chat;

    public RecipeRunner(@NotNull Project project, @NotNull UpdatableChat chat) {

        this.project = project;
        this.chat = chat;
    }

    private String getMarkdownFormatPrompt() {
        return "Enclose code snippets with three backticks like so: ```.";
    }

    public void runExplainCodeDetailed() {
        EditorContext editorContext = EditorContextGetter.getEditorContext(project);
        if (editorContext.getSelection() == null) {
            chat.addMessage(ChatMessage.createAssistantMessage("No code selected. Please select some code and try again."));
            return;
        }
        String languageName = LanguageUtils.getNormalizedLanguageName(editorContext.getCurrentFileExtension());

        String truncatedSelectedText = TruncationUtils.truncateText(editorContext.getSelection(), TruncationUtils.MAX_RECIPE_INPUT_TOKENS);
        String truncatedPrecedingText = editorContext.getPrecedingText() != null ? TruncationUtils.truncateTextStart(editorContext.getPrecedingText(), TruncationUtils.MAX_RECIPE_SURROUNDING_TOKENS) : "";
        String truncatedFollowingText = editorContext.getFollowingText() != null ? TruncationUtils.truncateText(editorContext.getFollowingText(), TruncationUtils.MAX_RECIPE_SURROUNDING_TOKENS) : "";

        String promptMessage = String.format(
            "Please explain the following %s code. Be very detailed and specific, and indicate when it is not clear to you what is going on. Format your response as an ordered list.\n```\n%s\n```\n%s",
            languageName,
            truncatedSelectedText,
            getMarkdownFormatPrompt()
        );

        String displayText = String.format(
            "Explain the following code:\n```\n%s\n```",
            editorContext.getSelection()
        );



//        return new Interaction(
//            { speaker: 'human', text: promptMessage, displayText },
//        { speaker: 'assistant' },
//        getContextMessagesFromSelection(
//            truncatedSelectedText,
//            truncatedPrecedingText,
//            truncatedFollowingText,
//            selection.fileName,
//            context.codebaseContext
//        )
        ChatMessage humanMessage = ChatMessage.createHumanMessage(promptMessage, new ArrayList<>());

        chat.addMessage(humanMessage);
    }

//    private ArrayList<ChatMessage> getContextMessagesFromSelection(EditorContext editorContext) {
//        return ChatMessage.createHumanMessage(editorContext, new ArrayList<String>());
//    }

    public void runExplainCodeHighLevel() {

    }

    public void runGenerateUnitTest() {

    }

    public void runGenerateDocstring() {

    }

    public void runImproveVariableNames() {

    }

    public void runTranslateToLanguage() {

    }

    public void runGitHistory() {

    }

    public void runFindCodeSmells() {

    }

    public void runFixup() {

    }

    public void runContextSearch() {

    }

    public void runReleaseNotes() {

    }
}
