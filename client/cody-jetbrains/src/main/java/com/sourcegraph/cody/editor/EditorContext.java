package com.sourcegraph.cody.editor;

import org.jetbrains.annotations.Nullable;

public class EditorContext {
    @Nullable
    private final String currentFileName;
    @Nullable
    private final String currentFileContent;
    @Nullable
    private final String selection;

    public EditorContext(@Nullable String currentFileName, @Nullable String currentFileContent, @Nullable String selection) {
        this.currentFileName = currentFileName;
        this.currentFileContent = currentFileContent;
        this.selection = selection;
    }

    @Nullable
    public String getCurrentFileName() {
        return currentFileName;
    }

    @Nullable
    public String getCurrentFileContent() {
        return currentFileContent;
    }

    @Nullable
    public String getSelection() {
        return selection;
    }
}
