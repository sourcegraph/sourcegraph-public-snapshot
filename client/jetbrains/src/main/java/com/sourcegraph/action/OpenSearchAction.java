package com.sourcegraph.action;

import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.application.AppUIExecutor;
import com.intellij.openapi.project.DumbAware;
import com.intellij.openapi.project.ProjectManager;
import com.sourcegraph.ui.SourcegraphWindow;
import org.jetbrains.annotations.NotNull;

public class OpenSearchAction extends AnAction implements DumbAware {
    private SourcegraphWindow window;

    @Override
    public void actionPerformed(@NotNull AnActionEvent e) {
        AppUIExecutor.onUiThread().execute(() -> {
            if (this.window == null) {
                this.window = new SourcegraphWindow(e.getProject());
            }
            this.window.showPopup();
        });
    }
}
