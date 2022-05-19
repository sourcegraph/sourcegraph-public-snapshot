package com.sourcegraph.find;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.popup.JBPopup;
import com.intellij.openapi.ui.popup.JBPopupFactory;
import com.intellij.openapi.util.Disposer;
import org.jetbrains.annotations.NotNull;

public class SourcegraphWindow implements Disposable {
    private final Project project;
    private final FindPopupPanel mainPanel;
    private JBPopup popup;

    public SourcegraphWindow(@NotNull Project project) {
        this.project = project;

        // Create main panel
        mainPanel = new FindPopupPanel(project);

        Disposer.register(project, this);
    }

    synchronized public void showPopup() {
        if (popup == null || popup.isDisposed()) {
            popup = createPopup();
            popup.showCenteredInCurrentWindow(project);
        }

        // If the popup is already shown, hitting alt + a gain should behave the same as the native find in files
        // feature and focus the search field.
        if (mainPanel.getBrowser() != null) {
            mainPanel.getBrowser().focus();
        }
    }

    @NotNull
    private JBPopup createPopup() {
        return JBPopupFactory.getInstance().createComponentPopupBuilder(mainPanel, mainPanel)
            .setTitle("Find on Sourcegraph")
            .setCancelOnClickOutside(true)
            .setResizable(true)
            .setModalContext(false)
            .setRequestFocus(true)
            .setFocusable(true)
            .setMovable(true)
            .setBelongsToGlobalPopupStack(true)
            .setCancelOnOtherWindowOpen(true)
            .setCancelKeyEnabled(true)
            .setNormalWindowLevel(true)
            .createPopup();
    }

    @Override
    public void dispose() {
        if (popup != null) {
            popup.dispose();
        }
    }
}
