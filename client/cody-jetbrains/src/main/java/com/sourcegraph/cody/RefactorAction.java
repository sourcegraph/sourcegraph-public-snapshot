package com.sourcegraph.cody;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.sourcegraph.completions.CompletionsInput;
import com.sourcegraph.completions.CompletionsService;
import com.sourcegraph.completions.Speaker;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.util.ArrayList;

public class RefactorAction extends DumbAwareAction { // TODO: Link this action in plugin.xml
    @Override
    public void actionPerformed(@NotNull AnActionEvent event) {
        Project project = event.getProject();
        if (project == null) {
            return;
        }
        EditorContext editorContext = EditorContextGetter.getEditorContext(project);
        if (editorContext == null) { // TODO: Don't fail if there's no editor context
            return;
        }
        try {
            // TODO: Use the prompt from VS Code
            var codeContext = "";
            if (editorContext.getSelection() == null) {
                codeContext = "Here is my current file\n" + editorContext.getCurrentFileContent();
            } else {
                codeContext = "Here is my highlighted code\n" + editorContext.getSelection();
            }
            String prompt = JOptionPane.showInputDialog("Prompt:");
            String context = "You are a software engineer who can write code according to instructions. You are given some code for extra context. You must not say anything other than the full code. You must output the full code snippet, and only the code snippet. You must not wrap the code snippet with markdown. You must double check you do not say anything other than the code.";

            if (StringUtils.isEmpty(prompt)) {
                return;
            }

            var input = new CompletionsInput(new ArrayList<>(), 0.5f, 1000, -1, -1);
            input.addMessage(Speaker.HUMAN, context);
            input.addMessage(Speaker.ASSISTANT, "Ok.");
            input.addMessage(Speaker.HUMAN, codeContext);
            input.addMessage(Speaker.ASSISTANT, "Ok.");
            input.addMessage(Speaker.HUMAN, prompt);
            input.addMessage(Speaker.ASSISTANT, "");

            input.messages.forEach(System.out::println);

            // ConfigUtil.getAccessToken(project)
            String result = new CompletionsService("http://localhost:3080/.api/graphql", "").send(input);
            System.out.println(result);
            // TODO: Use output

        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
