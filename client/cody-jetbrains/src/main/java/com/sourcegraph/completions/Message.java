package com.sourcegraph.completions;

import org.jetbrains.annotations.NotNull;

public class Message {
    @NotNull
    private final Speaker speaker;
    @NotNull
    private final String text;

    public Message(@NotNull Speaker speaker, @NotNull String text) {
        this.speaker = speaker;
        this.text = text;
    }

    @Override
    public String toString() {
        return String.format("Message { speaker=%s, text='%s'}", speaker, text);
    }
}
