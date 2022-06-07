package com.sourcegraph.find;

import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.intellij.codeInsight.daemon.DaemonCodeAnalyzer;
import com.intellij.openapi.externalSystem.service.execution.NotSupportedException;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileEditor.OpenFileDescriptor;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.PsiFile;
import com.intellij.psi.PsiManager;
import com.intellij.testFramework.LightVirtualFile;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.awt.*;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.Base64;
import java.util.Objects;

public class PreviewContent {
    private final Project project;
    private final String fileName;
    private final String repoUrl;
    private final String path;
    private final String content;
    private final int lineNumber;
    private final int[][] absoluteOffsetAndLengths;
    private final String relativeUrl;

    private VirtualFile virtualFile;

    public PreviewContent(@NotNull Project project, @Nullable String fileName, @NotNull String repoUrl, @Nullable String path, @Nullable String content, int lineNumber, int[][] absoluteOffsetAndLengths, @Nullable String relativeUrl) {
        this.project = project;
        // It seems like the constructor is not called when we use the JSON parser to create instances of this class, so
        // avoid adding any computation here.
        this.fileName = fileName;
        this.repoUrl = repoUrl;
        this.path = path;
        this.content = content;
        this.lineNumber = lineNumber;
        this.absoluteOffsetAndLengths = absoluteOffsetAndLengths;
        this.relativeUrl = relativeUrl;
    }

    public static PreviewContent fromJson(Project project, JsonObject json) {
        int absoluteOffsetAndLengthsSize = json.getAsJsonArray("absoluteOffsetAndLengths").size();
        int[][] absoluteOffsetAndLengths = new int[absoluteOffsetAndLengthsSize][2];
        for (int i = 0; i < absoluteOffsetAndLengths.length; i++) {
            JsonElement element = json.getAsJsonArray("absoluteOffsetAndLengths").get(i);
            absoluteOffsetAndLengths[i][0] = element.getAsJsonArray().get(0).getAsInt();
            absoluteOffsetAndLengths[i][1] = element.getAsJsonArray().get(1).getAsInt();
        }

        return new PreviewContent(project,
            json.get("fileName").getAsString(),
            json.get("repoUrl").getAsString(),
            json.get("path").getAsString(),
            json.get("content").getAsString(),
            json.get("lineNumber").getAsInt(),
            absoluteOffsetAndLengths,
            json.get("relativeUrl").getAsString());
    }

    @Nullable
    public String getFileName() {
        return fileName;
    }

    @Nullable
    public String getRepoUrl() {
        return repoUrl;
    }

    @Nullable
    public String getPath() {
        return path;
    }

    @Nullable
    public String getContent() {
        return convertBase64ToString(content);
    }

    public int getLineNumber() {
        return lineNumber;
    }

    public int[][] getAbsoluteOffsetAndLengths() {
        return absoluteOffsetAndLengths;
    }

    @Nullable
    public String getRelativeUrl() {
        return relativeUrl;
    }

    @Nullable
    public VirtualFile getVirtualFile() {
        if (virtualFile == null && fileName != null && content != null) {
            virtualFile = new LightVirtualFile(fileName, content);
        }
        return virtualFile;
    }

    @Nullable
    private static String convertBase64ToString(@Nullable String base64String) {
        if (base64String == null) {
            return null;
        }
        byte[] decodedBytes = Base64.getDecoder().decode(base64String);
        return new String(decodedBytes);
    }

    @Override
    public boolean equals(@Nullable Object obj) {
        return obj instanceof PreviewContent && equals((PreviewContent) obj);
    }

    private boolean equals(PreviewContent other) {
        return Objects.equals(fileName, other.fileName)
            && repoUrl.equals(other.repoUrl)
            && Objects.equals(path, other.path)
            && Objects.equals(content, other.content)
            && lineNumber == other.lineNumber
            && Objects.deepEquals(absoluteOffsetAndLengths, other.absoluteOffsetAndLengths)
            && Objects.equals(relativeUrl, other.relativeUrl);
    }

    public void openInEditorOrBrowser() throws URISyntaxException, IOException, NotSupportedException {
        if (fileName == null || fileName.length() == 0) {
            openInBrowser();
        } else {
            openInEditor();
        }
    }

    private void openInEditor() {
        assert fileName != null;
        // Open file in editor
        virtualFile = new LightVirtualFile(fileName, Objects.requireNonNull(content));
        OpenFileDescriptor openFileDescriptor = new OpenFileDescriptor(project, virtualFile, 0);
        FileEditorManager.getInstance(project).openTextEditor(openFileDescriptor, true);

        // Suppress code issues
        PsiFile file = PsiManager.getInstance(project).findFile(virtualFile);
        if (file != null) {
            DaemonCodeAnalyzer.getInstance(project).setHighlightingEnabled(file, false);
        }
    }

    private void openInBrowser() throws URISyntaxException, IOException, NotSupportedException {
        // Source: https://stackoverflow.com/questions/5226212/how-to-open-the-default-webbrowser-using-java
        if (Desktop.isDesktopSupported() && Desktop.getDesktop().isSupported(Desktop.Action.BROWSE)) {
            String sourcegraphUrl = ConfigUtil.getSourcegraphUrl(project);
            Desktop.getDesktop().browse(new URI(sourcegraphUrl + "/" + relativeUrl));
        } else {
            throw new NotSupportedException("Can't open link. Desktop is not supported.");
        }
    }
}
