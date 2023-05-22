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
    private final String precedingText;
    @Nullable
    private final String selection;
    @Nullable
    private final String followingText;

    public EditorContext() {
        this(null, null, null, null,null);
    }

    public EditorContext(@Nullable String currentFileName, @Nullable String currentFileContent, @Nullable String precedingText, @Nullable String selection, @Nullable String followingText) {
        this.currentFileName = currentFileName;
        this.currentFileContent = currentFileContent;
        this.precedingText = precedingText;
        this.selection = selection;
        this.followingText = followingText;
    }

    public @Nullable String getCurrentFileName() {
        return currentFileName;
    }

    public @Nullable String getCurrentFileExtension() {
        return currentFileName != null ? currentFileName.substring(currentFileName.lastIndexOf(".") + 1) : null;
    }

    public @Nullable String getCurrentFileContent() {
        return currentFileContent;
    }

    public @Nullable String getPrecedingText() {
        return precedingText;
    }

    public @Nullable String getSelection() {
        return selection;
    }

    public @Nullable String getFollowingText() {
        return followingText;
    }

    public @NotNull ArrayList<String> getCurrentFileContentAsArrayList() {
        return currentFileContent != null ? new ArrayList<>(Collections.singletonList(this.getCurrentFileContent())) : new ArrayList<>();
    }
}
