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
    private boolean isFirstRender = true;

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

        this.recreateBrowserIfNeeded();

        // If the popup is already shown, hitting alt + a gain should behave the same as the native find in files
        // feature and focus the search field.
        if (mainPanel.getBrowser() != null) {
            mainPanel.getBrowser().focus();
        }
    }

    /**
     * This is a workaround for #34773: On Mac OS, the web view is empty after opening and closing the popover
     * repeatedly.
     *
     * We work around the issue by forcing a recreation of the JCEF browser window whenever we open the popover. This
     * increases the modal opening times drastically and adds noticeable lag.
     */
    private void recreateBrowserIfNeeded() {
        boolean isMacOS = System.getProperty("os.name").equals("Mac OS X");

        if (!isFirstRender && isMacOS) {
            mainPanel.createBrowserPanel();
        }

        isFirstRender = false;
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
