package com.sourcegraph.cody;

import com.intellij.openapi.project.Project;
import org.jetbrains.annotations.NotNull;

public class RecipeRunner {
    private final @NotNull Project project;
    private final @NotNull UpdatableChat chat;

    public RecipeRunner(@NotNull Project project, @NotNull UpdatableChat chat) {

        this.project = project;
        this.chat = chat;
    }

    public void runExplainCodeDetailed() {

    }

    public void runExplainCodeHighLevel() {

    }

    public void runGenerateUnitTest() {

    }

    public void runGenerateDocstring() {

    }

    public void runImproveVariableNames() {

    }

    public void runTranslateToLanguage() {

    }

    public void runGitHistory() {

    }

    public void runFindCodeSmells() {

    }

    public void runFixup() {

    }

    public void runContextSearch() {

    }

    public void runReleaseNotes() {

    }
}
