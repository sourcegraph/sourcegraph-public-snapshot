package com.sourcegraph.cody.chat;

import com.sourcegraph.cody.completions.CompletionsInput;
import com.sourcegraph.cody.completions.CompletionsService;
import com.sourcegraph.cody.completions.Speaker;
import org.jetbrains.annotations.NotNull;

import java.io.IOException;
import java.util.ArrayList;

public class Chat {
    public @NotNull ChatMessage sendMessage(@NotNull ChatMessage humanMessage) throws IOException, InterruptedException {
        // TODO: Use the prompt from VS Code
        var codeContext = "";
        if (humanMessage.getContextFiles().size() == 0) {
            codeContext = "I have no file open in the editor right now.";
        } else {
            codeContext = "Here is my current file\n" + humanMessage.getContextFiles().get(0);
        }
        String preamble = "You are a software engineer who can write code according to instructions. You are given some code for extra context. You must not say anything other than the full code. You must output the full code snippet, and only the code snippet. You must not wrap the code snippet with markdown. You must double check you do not say anything other than the code.";

        var input = new CompletionsInput(new ArrayList<>(), 0.5f, 1000, -1, -1);
        input.addMessage(Speaker.HUMAN, preamble);
        input.addMessage(Speaker.ASSISTANT, "Ok.");
        input.addMessage(Speaker.HUMAN, codeContext);
        input.addMessage(Speaker.ASSISTANT, "Ok.");
        input.addMessage(Speaker.HUMAN, humanMessage.getText());
        input.addMessage(Speaker.ASSISTANT, "");

        input.messages.forEach(System.out::println);

        // ConfigUtil.getAccessToken(project) TODO: Get the access token from the plugin config
        String result = new CompletionsService("http://localhost:3080/.api/graphql", "").getCompletion(input); // TODO: Don't create this each time

        return ChatMessage.createAssistantMessage(result);
    }
}
