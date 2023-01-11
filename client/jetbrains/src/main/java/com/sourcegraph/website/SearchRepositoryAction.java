package com.sourcegraph.website;

import com.intellij.openapi.actionSystem.AnActionEvent;
import org.jetbrains.annotations.NotNull;


@SuppressWarnings("MissingActionUpdateThread")
public class SearchRepositoryAction extends SearchActionBase {
    @Override
    public void actionPerformed(@NotNull AnActionEvent event) {
        super.actionPerformedMode(event, Scope.REPOSITORY);
    }
}
