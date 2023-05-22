package com.sourcegraph.cody.context;

import com.sourcegraph.cody.completions.Message;
import com.sourcegraph.cody.completions.Speaker;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class ContextMessage extends Message {
    @Nullable
    private final ContextFile file;

    public ContextMessage(@NotNull Speaker speaker, @NotNull String text, @Nullable ContextFile file) {
        super(speaker, text);
        this.file = file;
    }

    public @Nullable ContextFile getFile() {
        return file;
    }

    public static @NotNull ContextMessage createHumanMessage(@NotNull String text, @NotNull ContextFile file) {
        return new ContextMessage(Speaker.HUMAN, text, file);
    }

    public static @NotNull ContextMessage createDefaultAssistantMessage() {
        return new ContextMessage(Speaker.ASSISTANT, "Ok.", null);
    }
}
