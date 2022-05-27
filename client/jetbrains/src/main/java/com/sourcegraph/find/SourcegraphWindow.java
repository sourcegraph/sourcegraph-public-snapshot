package com.sourcegraph.find;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.ui.popup.ComponentPopupBuilder;
import com.intellij.openapi.ui.popup.JBPopup;
import com.intellij.openapi.ui.popup.JBPopupFactory;
import com.intellij.openapi.util.Disposer;
import org.cef.browser.CefBrowser;
import org.cef.handler.CefKeyboardHandler;
import org.cef.misc.BoolRef;
import org.jetbrains.annotations.NotNull;

import java.awt.event.KeyAdapter;
import java.awt.event.KeyEvent;

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

        if (shouldHideInsteadOfCancel()) {
            popup.setUiVisible(true);
        }

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
     * We work around the issue by forcing hiding the Popover instead of clearing it on Mac OS. This slightly increases
     * resources consumption but allows us to reuse the JCEF window on this platform.
     */
    private boolean shouldHideInsteadOfCancel() {
        return System.getProperty("os.name").equals("Mac OS X");
    }

    @NotNull
    private JBPopup createPopup() {
        ComponentPopupBuilder builder = JBPopupFactory.getInstance().createComponentPopupBuilder(mainPanel, mainPanel)
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
            .setNormalWindowLevel(true);

        if (shouldHideInsteadOfCancel()) {
            builder = builder.setCancelCallback(() -> {
                popup.setUiVisible(false);
                // We return false to prevent the default cancellation behavior.
                return false;
            });

            // For some reason, adding a cancelCallback will prevent the cancel event to fire when using the escape
            // key. To work around this, we add a manual listener to both the popup panel and the browser panel for this
            // scenario.
            mainPanel.addKeyListener(new KeyAdapter() {
                public void keyPressed(KeyEvent event) {
                    if (event.getKeyCode() == KeyEvent.VK_ESCAPE) {
                        popup.setUiVisible(false);
                    }
                }
            });
            mainPanel.getBrowser().getJBCefClient().addKeyboardHandler(new CefKeyboardHandler() {
                @Override
                public boolean onPreKeyEvent(CefBrowser browser, CefKeyEvent event, BoolRef is_keyboard_shortcut) {
                    return false;
                }
                @Override
                public boolean onKeyEvent(CefBrowser browser, CefKeyEvent event) {
                    if (event.windows_key_code == KeyEvent.VK_ESCAPE) {
                        popup.setUiVisible(false);
                    }
                    return false;
                }
            }, mainPanel.getBrowser().getCefBrowser());
        }

        return builder.createPopup();
    }

    @Override
    public void dispose() {
        if (popup != null) {
            popup.dispose();
        }
    }
}
