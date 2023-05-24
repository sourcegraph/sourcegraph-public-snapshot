package com.sourcegraph.cody.prompts;

import org.jetbrains.annotations.NotNull;

public class Prompter {

    private static final @NotNull String CURRENT_EDITOR_CODE_TEMPLATE = "I have the `{filePath}` file opened in my editor. ";
    private static final @NotNull String CURRENT_EDITOR_SELECTED_CODE_TEMPLATE = "I am currently looking at this part of the code from `{filePath}`. ";
    private static final @NotNull String CODE_CONTEXT_TEMPLATE = "Use following code snippet from file `{filePath}`:\n```{language}\n{text}\n```";
    private static final @NotNull String TEXT_CONTEXT_TEMPLATE = "Use the following text from file `{filePath}`:\n{text}";

    public static @NotNull String getCurrentEditorCodePrompt(@NotNull String filePath, @NotNull String code) {
        String context = LanguageUtils.isMarkdownFile(filePath) ? getTextContextPrompt(filePath, code) : getCodeContextPrompt(filePath, code);
        return CURRENT_EDITOR_CODE_TEMPLATE.replace("{filePath}", filePath) + context;
    }

    public static @NotNull String getCurrentEditorSelectedCode(@NotNull String filePath, @NotNull String code) {
        String context = LanguageUtils.isMarkdownFile(filePath) ? getTextContextPrompt(filePath, code) : getCodeContextPrompt(filePath, code);
        return CURRENT_EDITOR_SELECTED_CODE_TEMPLATE.replace("{filePath}", filePath) + context;
    }

    public static @NotNull String getContextPrompt(@NotNull String filePath, @NotNull String code) {
        return LanguageUtils.isMarkdownFile(filePath) ? getTextContextPrompt(filePath, code) : getCodeContextPrompt(filePath, code);
    }

    public static @NotNull String getCodeContextPrompt(@NotNull String filePath, @NotNull String code) {
        return CODE_CONTEXT_TEMPLATE.replace("{filePath}", filePath)
            .replace("{language}", LanguageUtils.getNormalizedLanguageName(filePath))
            .replace("{text}", code);
    }

    public static @NotNull String getTextContextPrompt(@NotNull String filePath, @NotNull String text) {
        return TEXT_CONTEXT_TEMPLATE.replace("{filePath}", filePath).replace("{text}", text);
    }
}
