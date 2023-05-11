package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.completions.CompletionsInput;
import com.sourcegraph.cody.completions.CompletionsService;
import com.sourcegraph.cody.completions.Message;
import com.sourcegraph.cody.completions.Speaker;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

public class Chat {
    private final @Nullable String codebase;
    private final @NotNull String instanceUrl;
    private final @NotNull String accessToken;

    public Chat(@Nullable String codebase, @NotNull String instanceUrl, @NotNull String accessToken) {
        this.codebase = codebase;
        this.instanceUrl = instanceUrl;
        this.accessToken = accessToken;
    }

    public @NotNull ChatMessage sendMessage(@NotNull ChatMessage humanMessage) throws IOException, InterruptedException {
        List<Message> preamble = Preamble.getPreamble(codebase);

        // TODO: Use the context getting logic from VS Code
        var codeContext = "";
        if (humanMessage.getContextFiles().size() == 0) {
            codeContext = "I have no file open in the editor right now.";
        } else {
            codeContext = "Here is my current file\n" + humanMessage.getContextFiles().get(0);
        }

        var input = new CompletionsInput(new ArrayList<>(), 0.5f, 1000, -1, -1);
        input.addMessages(preamble);
        input.addMessage(Speaker.HUMAN, codeContext);
        input.addMessage(Speaker.ASSISTANT, "Ok.");
        input.addMessage(Speaker.HUMAN, humanMessage.getText());
        input.addMessage(Speaker.ASSISTANT, "");

        input.messages.forEach(System.out::println);

        // ConfigUtil.getAccessToken(project) TODO: Get the access token from the plugin config
        String result = new CompletionsService(instanceUrl, accessToken).getCompletion(input); // TODO: Don't create this each time

        return ChatMessage.createAssistantMessage(result);
    }
}
