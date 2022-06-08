package com.sourcegraph.find;

import org.jetbrains.annotations.Nullable;

import java.util.Base64;

public class PreviewContent {
    private final String fileName;
    private final String path;
    private final String content;
    private final int lineNumber;
    private final int[][] absoluteOffsetAndLengths;
    private final String relativeUrl;

    public PreviewContent(String fileName, String path, String content, int lineNumber, int[][] absoluteOffsetAndLengths, String relativeUrl) {
        // It seems like the constructor is not called when we use the JSON parser to create instances of this class, so
        // avoid adding any computation here.
        this.fileName = fileName;
        this.path = path;
        this.content = content;
        this.lineNumber = lineNumber;
        this.absoluteOffsetAndLengths = absoluteOffsetAndLengths;
        this.relativeUrl = relativeUrl;
    }

    public String getFileName() {
        return fileName;
    }

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

    public String getRelativeUrl() {
        return relativeUrl;
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
        return fileName.equals(other.fileName)
            && path.equals(other.path)
            && content.equals(other.content)
            && lineNumber == other.lineNumber
            && relativeUrl.equals(other.relativeUrl);
    }
}
