package com.sourcegraph.vcs;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import org.jetbrains.annotations.NotNull;

public class RevisionContext {
    private final Project project;
    private final String revisionNumber; // Revision number or commit hash
    private final VirtualFile repoRoot;

    public RevisionContext(@NotNull Project project, @NotNull String revisionNumber, @NotNull VirtualFile repoRoot) {
        this.project = project;
        this.revisionNumber = revisionNumber;
        this.repoRoot = repoRoot;
    }

    @NotNull
    public Project getProject() {
        return project;
    }

    @NotNull
    public String getRevisionNumber() {
        return revisionNumber;
    }

    @NotNull
    public VirtualFile getRepoRoot() {
        return repoRoot;
    }
}
