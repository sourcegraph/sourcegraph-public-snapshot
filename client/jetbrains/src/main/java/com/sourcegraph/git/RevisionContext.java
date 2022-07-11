package com.sourcegraph.git;

import com.intellij.openapi.project.Project;
import org.jetbrains.annotations.NotNull;

public class RevisionContext {
    private final Project project;
    private final String revisionNumber; // Revision number or commit hash

    public RevisionContext(@NotNull Project project, @NotNull String revisionNumber) {
        this.project = project;
        this.revisionNumber = revisionNumber;
    }

    @NotNull
    public Project getProject() {
        return project;
    }

    @NotNull
    public String getRevisionNumber() {
        return revisionNumber;
    }
}
