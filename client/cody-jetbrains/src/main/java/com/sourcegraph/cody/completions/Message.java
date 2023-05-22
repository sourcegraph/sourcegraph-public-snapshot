package com.sourcegraph.cody.completions;

import org.jetbrains.annotations.NotNull;

public class Message {
    protected final @NotNull Speaker speaker;
    protected final @NotNull String text;

    public Message(@NotNull Speaker speaker, @NotNull String text) {
        this.speaker = speaker;
        this.text = text;
    }

    public @NotNull Speaker getSpeaker() {
        return speaker;
    }

    public @NotNull String getText() {
        return text;
    }

    @Override
    public @NotNull String toString() {
        return String.format("Message { speaker=%s, text='%s'}", speaker, text);
    }
}
