package com.sourcegraph.find;

public class PreviewContent {
    private final String fileName;
    private final String path;
    private final String content;
    private final int lineNumber;
    private final int[][] absoluteOffsetAndLengths;

    public PreviewContent(String fileName, String path, String content, int lineNumber, int[][] absoluteOffsetAndLengths) {
        this.fileName = fileName;
        this.path = path;
        this.content = content;
        this.lineNumber = lineNumber;
        this.absoluteOffsetAndLengths = absoluteOffsetAndLengths;
    }

    public String getFileName() {
        return fileName;
    }

    public String getPath() {
        return path;
    }

    public String getContent() {
        return content;
    }

    public int getLineNumber() {
        return lineNumber;
    }

    public int[][] getAbsoluteOffsetAndLengths() {
        return absoluteOffsetAndLengths;
    }
}
