package com.sourcegraph.jody;

import com.intellij.codeInsight.actions.ReformatCodeAction;
import com.intellij.codeInsight.documentation.DocumentationActionProvider;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.CommonDataKeys;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.command.WriteCommandAction;
import com.intellij.openapi.editor.*;
import com.intellij.openapi.fileEditor.FileDocumentManager;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.project.DumbAwareAction;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.TextRange;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.PsiClass;
import com.intellij.psi.PsiElement;
import com.intellij.psi.PsiFile;
import com.intellij.psi.PsiMethod;
import com.intellij.psi.util.PsiTreeUtil;
import com.sourcegraph.config.ConfigUtil;
import com.theokanning.openai.completion.chat.ChatCompletionChoice;
import com.theokanning.openai.completion.chat.ChatCompletionRequest;
import com.theokanning.openai.completion.chat.ChatCompletionResult;
import com.theokanning.openai.completion.chat.ChatMessage;
import com.theokanning.openai.service.OpenAiService;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Consumer;
import java.util.stream.Collectors;

public class RefactorAction extends DumbAwareAction {

    @Override
    public void actionPerformed(@NotNull AnActionEvent event) {
        final Project project = event.getProject();
        if (project == null) {
            return;
        }
        Editor editor = FileEditorManager.getInstance(project).getSelectedTextEditor();
        if (editor == null) {
            return;
        }
        Document currentDocument = editor.getDocument();
        VirtualFile currentFile = FileDocumentManager.getInstance().getFile(currentDocument);
        if (currentFile == null) {
            return;
        }
        String selection = editor.getSelectionModel().getSelectedText();
        try {
            // I am here

             var codeContext = "";
//            if (selection == null || selection.isEmpty()) {
            codeContext = "Here is my current file\n" + currentDocument.getText();
//            } else {
//                codeContext = "Here is my highlighted code\n" + selection;
//            }
            String prompt = JOptionPane.showInputDialog("Prompt:");
            String context = "You are a software engineer who can write code according to instructions. You are given some code for extra context. You must not say anything other than the full code. You must output the full code snippet, and only the code snippet. You must not wrap the code snippet with markdown. You must double check you do not say anything other than the code.";

            if (StringUtils.isEmpty(prompt)) {
                return;
            }
//            var line = editor.getCaretModel().getCurrentCaret().getVisualPosition().line;

            PsiFile psiFile = event.getData(CommonDataKeys.PSI_FILE);
            PsiElement elementAt = psiFile.findElementAt(editor.getCaretModel().getOffset());
            PsiElement currentline = PsiTreeUtil.getStubOrPsiParent(elementAt);
            PsiElement psiClass = PsiTreeUtil.getStubOrPsiParentOfType(elementAt, PsiClass.class);
            PsiElement psiMethod = PsiTreeUtil.getStubOrPsiParentOfType(elementAt, PsiMethod.class);
            PsiElement psiblock = elementAt.getContext();
//            System.out.println(currentline.toString());
//            System.out.println(currentline.getText());
//            System.out.println("asdf");

            PsiElement parent = PsiTreeUtil.getStubOrPsiParent(currentline);
//            System.out.println(parent.getText());

//            var line = currentline.getText();


            var offset = editor.getCaretModel().getOffset();
            var lineNum = currentDocument.getLineNumber(offset);
            var textRange = new TextRange(currentDocument.getLineStartOffset(lineNum), currentDocument.getLineEndOffset(lineNum));
            var line = currentDocument.getText(textRange);

            CompletionsInput.CompletionsInputBuilder builder = CompletionsInput.builder()
                .addMessage(SpeakerType.HUMAN, context)
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, codeContext)
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, "the current line is: " + line)
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, "the current class is: " + psiClass.getText())
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, psiMethod == null ? "You are not currently in a method." : "the current method is: " + psiMethod.getText())
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, "the current code block is: " + psiblock.getText())
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, prompt)
                .addMessage(SpeakerType.ASSISTANT, "")
                 .setTemperature(0.5f)
                .setTopK(-1)
                .setTopP(-1)
                .setMaxTokensToSample(1000);



            var input = builder.build();

            System.out.println("context:");
            input.getMessages().forEach(System.out::println);

            String result = new CompletionsService("http://localhost:3080/.api/graphql", ConfigUtil.getAccessToken(project)).send(input);
            System.out.println(result);
            WriteCommandAction.runWriteCommandAction(project, () -> currentDocument.replaceString(editor.getSelectionModel().getSelectionStart(), editor.getSelectionModel().getSelectionEnd(), result));

        } catch (Exception ex) {
//            ex.printStackTrace();

        }
    }

 private CompletionsInput generateCompletionsInput(Editor editor, PsiFile psiFile) {
        var offset = editor.getCaretModel().getOffset();
        var lineNum = editor.getDocument().getLineNumber(offset);
        var textRange = new TextRange(editor.getDocument().getLineStartOffset(lineNum), editor.getDocument().getLineEndOffset(lineNum));
        var line = editor.getDocument().getText(textRange);

        PsiElement elementAt = psiFile.findElementAt(offset);
        PsiElement currentline = PsiTreeUtil.getStubOrPsiParent(elementAt);
        PsiElement psiClass = PsiTreeUtil.getStubOrPsiParentOfType(elementAt, PsiClass.class);
        PsiElement psiMethod = PsiTreeUtil.getStubOrPsiParentOfType(elementAt, PsiMethod.class);
        PsiElement psiblock = elementAt.getContext();

        return CompletionsInput.builder()
                .addMessage(SpeakerType.HUMAN, "the current line is: " + line)
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, "the current class is: " + psiClass.getText())
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, psiMethod == null ? "You are not currently in a method." : "the current method is: " + psiMethod.getText())
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, "the current code block is: " + psiblock.getText())
                .addMessage(SpeakerType.ASSISTANT, "Ok.")
                .addMessage(SpeakerType.HUMAN, "Prompt:")
                .addMessage(SpeakerType.ASSISTANT, "")
                .setTemperature(0.5f)
                .setTopK(-1)
                .setTopP(-1)
                .setMaxTokensToSample(1000)
                .build();
    }

}


