package com.sourcegraph.action;

import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.application.AppUIExecutor;
import com.intellij.openapi.project.DumbAware;
import com.sourcegraph.ui.SourcegraphWindow;
import org.jetbrains.annotations.NotNull;

import java.util.Objects;

public class OpenSearchAction extends AnAction implements DumbAware {
    private SourcegraphWindow window;

    @Override
    public void actionPerformed(@NotNull AnActionEvent event) {
        AppUIExecutor.onUiThread().execute(() -> { // TODO: .execute() is not part of the public API, use something else instead
            if (this.window == null) {
                this.window = new SourcegraphWindow(Objects.requireNonNull(event.getProject()));
            }
            this.window.showPopup();
        });
    }
}
