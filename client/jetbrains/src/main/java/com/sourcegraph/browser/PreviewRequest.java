package com.sourcegraph.browser;

public class PreviewRequest {
    private final String fileName;
    private final String path;
    private final String content;
    private final String lineNumber;

    public PreviewRequest(String fileName, String path, String content, String lineNumber) {
        this.fileName = fileName;
        this.path = path;
        this.content = content;
        this.lineNumber = lineNumber;
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

    public String getLineNumber() {
        return lineNumber;
    }
}
