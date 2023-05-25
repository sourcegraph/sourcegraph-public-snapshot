package com.sourcegraph.cody.prompts;

import com.sourcegraph.cody.completions.Message;
import com.sourcegraph.cody.completions.Speaker;
import org.jetbrains.annotations.Nullable;

import java.util.ArrayList;
import java.util.List;

public class Preamble {

    private static final String actions = "You are Cody, an AI-powered coding assistant created by Sourcegraph. You work inside a text editor. You have access to my currently open files. You perform the following actions:\n" +
        "- Answer general programming questions.\n" +
        "- Answer questions about the code that I have provided to you.\n" +
        "- Generate code that matches a written description.\n" +
        "- Explain what a section of code does.";

    private static final String rules = "In your responses, obey the following rules:\n" +
        "- Be as brief and concise as possible without losing clarity.\n" +
        "- All code snippets have to be markdown-formatted, and placed in-between triple backticks like this ```.\n" +
        "- Answer questions only if you know the answer or can make a well-informed guess. Otherwise, tell me you don't know and what context I need to provide you for you to answer the question.\n" +
        "- Only reference file names or URLs if you are sure they exist.";

    private static final String answer = "Understood. I am Cody, an AI assistant made by Sourcegraph to help with programming tasks.\n" +
        "I work inside a text editor. I have access to your currently open files in the editor.\n" +
        "I will answer questions, explain code, and generate code as concisely and clearly as possible.\n" +
        "My responses will be formatted using Markdown syntax for code blocks.\n" +
        "I will acknowledge when I don't know an answer or need more context.";

    public static List<Message> getPreamble(@Nullable String codebase) {
        List<String> preamble = new ArrayList<>();
        preamble.add(actions);
        preamble.add(rules);

        List<String> preambleResponse = new ArrayList<>();
        preambleResponse.add(answer);

        // If we have a codebase, add a preamble about it
        if (codebase != null) {
            String codebasePreamble = "You have access to the `" + codebase + "` repository. You are able to answer questions about the `" + codebase + "` repository. " +
                "I will provide the relevant code snippets from the `" + codebase + "` repository when necessary to answer my questions.";

            preamble.add(codebasePreamble);
            preambleResponse.add("I have access to the `" + codebase + "` repository and can answer questions about its files.");
        }

        // Return this as a list of two items
        List<Message> messages = new ArrayList<>();
        messages.add(new Message(Speaker.HUMAN, String.join("\n\n", preamble)));
        messages.add(new Message(Speaker.ASSISTANT, String.join("\n", preambleResponse)));
        return messages;
    }
}
