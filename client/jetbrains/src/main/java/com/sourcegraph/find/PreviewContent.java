package com.sourcegraph.find;

import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.intellij.codeInsight.daemon.DaemonCodeAnalyzer;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileEditor.OpenFileDescriptor;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.PsiFile;
import com.intellij.psi.PsiManager;
import com.sourcegraph.common.BrowserOpener;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
import java.util.Base64;
import java.util.Date;
import java.util.Objects;

public class PreviewContent {
    private final Project project;
    private final Date receivedDateTime;
    private final String resultType;
    private final String fileName;
    private final String repoUrl;
    private final String commit;
    private final String path;
    private final String content;
    private final String symbolName;
    private final String symbolContainerName;
    private final String commitMessagePreview;
    private final int lineNumber;
    private final int[][] absoluteOffsetAndLengths;
    private final String relativeUrl;

    private VirtualFile virtualFile;

    public PreviewContent(@NotNull Project project,
                          @NotNull Date receivedDateTime,
                          @Nullable String resultType,
                          @Nullable String fileName,
                          @NotNull String repoUrl,
                          @Nullable String commit,
                          @Nullable String path,
                          @Nullable String content,
                          @Nullable String symbolName,
                          @Nullable String symbolContainerName,
                          @Nullable String commitMessagePreview,
                          int lineNumber,
                          int[][] absoluteOffsetAndLengths,
                          @Nullable String relativeUrl) {
        this.project = project;
        // It seems like the constructor is not called when we use the JSON parser to create instances of this class, so
        // avoid adding any computation here.
        this.receivedDateTime = receivedDateTime;
        this.resultType = resultType;
        this.fileName = fileName;
        this.repoUrl = repoUrl;
        this.commit = commit;
        this.path = path;
        this.symbolName = symbolName;
        this.symbolContainerName = symbolContainerName;
        this.commitMessagePreview = commitMessagePreview;
        this.content = content;
        this.lineNumber = lineNumber;
        this.absoluteOffsetAndLengths = absoluteOffsetAndLengths;
        this.relativeUrl = relativeUrl;
    }

    @NotNull
    public static PreviewContent fromJson(Project project, @NotNull JsonObject json) {
        int absoluteOffsetAndLengthsSize = isNotNull(json, "absoluteOffsetAndLengths") ? json.getAsJsonArray("absoluteOffsetAndLengths").size() : 0;
        int[][] absoluteOffsetAndLengths = new int[absoluteOffsetAndLengthsSize][2];
        for (int i = 0; i < absoluteOffsetAndLengths.length; i++) {
            JsonElement element = json.getAsJsonArray("absoluteOffsetAndLengths").get(i);
            absoluteOffsetAndLengths[i][0] = element.getAsJsonArray().get(0).getAsInt();
            absoluteOffsetAndLengths[i][1] = element.getAsJsonArray().get(1).getAsInt();
        }

        return new PreviewContent(project,
            Date.from(Instant.from(DateTimeFormatter.ISO_INSTANT.parse(json.get("timeAsISOString").getAsString()))),
            isNotNull(json, "resultType") ? json.get("resultType").getAsString() : null,
            isNotNull(json, "fileName") ? json.get("fileName").getAsString() : null,
            json.get("repoUrl").getAsString(),
            isNotNull(json, "commit") ? json.get("commit").getAsString() : null,
            isNotNull(json, "path") ? json.get("path").getAsString() : null,
            isNotNull(json, "content") ? json.get("content").getAsString() : null,
            isNotNull(json, "symbolName") ? json.get("symbolName").getAsString() : null,
            isNotNull(json, "symbolContainerName") ? json.get("symbolContainerName").getAsString() : null,
            isNotNull(json, "commitMessagePreview") ? json.get("commitMessagePreview").getAsString() : null,
            isNotNull(json, "lineNumber") ? json.get("lineNumber").getAsInt() : -1,
            absoluteOffsetAndLengths,
            isNotNull(json, "relativeUrl") ? json.get("relativeUrl").getAsString() : null);
    }

    private static boolean isNotNull(@NotNull JsonObject json, String key) {
        return json.get(key) != null && !json.get(key).isJsonNull();
    }

    @NotNull
    public Date getReceivedDateTime() {
        return receivedDateTime;
    }

    @Nullable
    public String getResultType() {
        return resultType;
    }

    @NotNull
    public String getRepoUrl() {
        return repoUrl;
    }

    @Nullable
    public String getCommit() {
        return commit;
    }

    @Nullable
    public String getPath() {
        return path;
    }

    @Nullable
    public String getContent() {
        return convertBase64ToString(content);
    }

    @Nullable
    public String getSymbolName() {
        return convertBase64ToString(symbolName);
    }

    @Nullable
    public String getSymbolContainerName() {
        return symbolContainerName;
    }

    @Nullable
    public String getCommitMessagePreview() {
        return commitMessagePreview;
    }

    public int[][] getAbsoluteOffsetAndLengths() {
        return absoluteOffsetAndLengths;
    }

    @NotNull
    public VirtualFile getVirtualFile() {
        if (virtualFile == null) {
            assert fileName != null; // We should always have a non-null file name and content when we call getVirtualFile()
            assert content != null;
            virtualFile = new SourcegraphVirtualFile(fileName, Objects.requireNonNull(getContent()), getRepoUrl(), getCommit(), getPath());
        }
        return virtualFile;
    }

    @Nullable
    private static String convertBase64ToString(@Nullable String base64String) {
        if (base64String == null) {
            return null;
        }
        byte[] decodedBytes = Base64.getDecoder().decode(base64String);
        return new String(decodedBytes, StandardCharsets.UTF_8);
    }

    @Override
    public boolean equals(@Nullable Object obj) {
        return obj instanceof PreviewContent && equals((PreviewContent) obj);
    }

    private boolean equals(@Nullable PreviewContent other) {
        return other != null && Objects.equals(fileName, other.fileName)
            && repoUrl.equals(other.repoUrl)
            && Objects.equals(path, other.path)
            && Objects.equals(content, other.content)
            && Objects.equals(symbolName, other.symbolName)
            && Objects.equals(symbolContainerName, other.symbolContainerName)
            && Objects.equals(commitMessagePreview, other.commitMessagePreview)
            && lineNumber == other.lineNumber
            && Objects.deepEquals(absoluteOffsetAndLengths, other.absoluteOffsetAndLengths)
            && Objects.equals(relativeUrl, other.relativeUrl);
    }

    public void openInEditorOrBrowser() {
        if (opensInEditor()) {
            openInEditor();
        } else {
            if (relativeUrl != null) {
                BrowserOpener.openRelativeUrlInBrowser(project, relativeUrl);
            }
        }
    }

    public boolean opensInEditor() {
        return fileName != null && fileName.length() > 0;
    }

    private void openInEditor() {
        assert fileName != null; // We should always have a non-null file name when we call openInEditor()
        // Open file in editor
        virtualFile = getVirtualFile();
        OpenFileDescriptor openFileDescriptor = new OpenFileDescriptor(project, virtualFile, 0);
        FileEditorManager.getInstance(project).openTextEditor(openFileDescriptor, true);

        // Suppress code issues
        PsiFile file = PsiManager.getInstance(project).findFile(virtualFile);
        if (file != null) {
            DaemonCodeAnalyzer.getInstance(project).setHighlightingEnabled(file, false);
        }
    }
}
