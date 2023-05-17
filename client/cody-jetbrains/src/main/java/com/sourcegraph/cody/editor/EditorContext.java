package com.sourcegraph.cody.editor;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.util.ArrayList;
import java.util.Collections;

public class EditorContext {
    @Nullable
    private final String currentFileName;
    @Nullable
    private final String currentFileContent;
    @Nullable
    private final String selection;

    public EditorContext() {
        this(null, null, null);
    }

    public EditorContext(@Nullable String currentFileName, @Nullable String currentFileContent, @Nullable String selection) {
        this.currentFileName = currentFileName;
        this.currentFileContent = currentFileContent;
        this.selection = selection;
    }

    public @Nullable String getCurrentFileName() {
        return currentFileName;
    }

    public @Nullable String getCurrentFileContent() {
        return currentFileContent;
    }

    public @Nullable String getSelection() {
        return selection;
    }

    public @NotNull ArrayList<String> getCurrentFileContentAsArrayList() {
        return currentFileContent != null ? new ArrayList<>(Collections.singletonList(this.getCurrentFileContent())) : new ArrayList<>();
    }
}
