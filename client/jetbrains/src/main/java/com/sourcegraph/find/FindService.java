package com.sourcegraph.find;

import com.intellij.ide.DataManager;
import com.intellij.openapi.Disposable;
import com.intellij.openapi.actionSystem.*;
import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.wm.ex.WindowManagerEx;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.find.browser.BrowserAndLoadingPanel;
import com.sourcegraph.find.browser.JavaToJSBridge;
import org.jetbrains.annotations.NotNull;

import java.awt.*;
import java.awt.event.KeyEvent;
import java.awt.event.WindowEvent;

import static java.awt.event.InputEvent.ALT_DOWN_MASK;
import static java.awt.event.WindowEvent.WINDOW_GAINED_FOCUS;

public class FindService implements Disposable {
    private final Project project;
    private final FindPopupPanel mainPanel;
    private FindPopupDialog popup;
    private static final Logger logger = Logger.getInstance(FindService.class);

    public FindService(@NotNull Project project) {
        this.project = project;

        // Create main panel
        mainPanel = new FindPopupPanel(project, this);
    }

    synchronized public void showPopup() {
        createOrShowPopup();
    }

    public void hidePopup() {
        popup.hide();
        hideMaterialUiOverlay();
    }

    private void createOrShowPopup() {
        if (popup != null) {
            if (!popup.isVisible()) {
                popup.show();

                // Retry auth and search if the popup is in a possible connection error state
                if (mainPanel.browserHasSearchError()
                    || mainPanel.getConnectionAndAuthState() == BrowserAndLoadingPanel.ConnectionAndAuthState.COULD_NOT_CONNECT) {
                    JavaToJSBridge bridge = mainPanel.getJavaToJSBridge();
                    if (bridge != null) {
                        bridge.callJS("retrySearch", null);
                    }
                }
            }
        } else {
            popup = new FindPopupDialog(project, mainPanel);

            // We add a manual listener to the global key handler since the editor component seems to work around the
            // default Swing event handler.
            registerGlobalKeyListeners();

            // We also need to detect when the main IDE frame or another popup inside the project gets focus and close
            // the Sourcegraph window accordingly.
            registerOutsideClickListener();
        }
    }

    private void registerGlobalKeyListeners() {
        KeyboardFocusManager.getCurrentKeyboardFocusManager()
            .addKeyEventDispatcher(e -> {
                if (e.getID() != KeyEvent.KEY_PRESSED || popup != null && (popup.isDisposed() || !popup.isVisible())) {
                    return false;
                }

                return handleKeyPress(e.getKeyCode(), e.getModifiersEx());
            });
    }

    private boolean handleKeyPress(int keyCode, int modifiers) {
        if (keyCode == KeyEvent.VK_ESCAPE && modifiers == 0) {
            ApplicationManager.getApplication().invokeLater(this::hidePopup);
            return true;
        }

        if (keyCode == KeyEvent.VK_ENTER && (modifiers & ALT_DOWN_MASK) == ALT_DOWN_MASK) {
            if (mainPanel.getPreviewPanel() != null && mainPanel.getPreviewPanel().getPreviewContent() != null) {
                // This must run on EDT (Event Dispatch Thread) because it may interact with the editor.
                ApplicationManager.getApplication().invokeLater(() -> {
                    try {
                        mainPanel.getPreviewPanel().getPreviewContent().openInEditorOrBrowser();
                    } catch (Exception e) {
                        logger.error("Error opening file in editor", e);
                    }
                });
                return true;
            }
        }

        return false;
    }

    private void registerOutsideClickListener() {
        Window projectParentWindow = getParentWindow(null);

        Toolkit.getDefaultToolkit().addAWTEventListener(event -> {
            if (event instanceof WindowEvent) {
                WindowEvent windowEvent = (WindowEvent) event;

                // We only care for focus events
                if (windowEvent.getID() != WINDOW_GAINED_FOCUS) {
                    return;
                }

                if (!this.popup.isVisible()) {
                    return;
                }

                // Detect if we're focusing the Sourcegraph popup
                if (windowEvent.getComponent().equals(this.popup.getWindow())) {
                    return;
                }

                // Detect if the newly focused window is a parent of the project root window
                Window currentProjectParentWindow = getParentWindow(windowEvent.getComponent());
                if (currentProjectParentWindow.equals(projectParentWindow)) {
                    hidePopup();
                }
            }
        }, AWTEvent.WINDOW_EVENT_MASK);
    }

    // https://sourcegraph.com/github.com/JetBrains/intellij-community@27fee7320a01c58309a742341dd61deae57c9005/-/blob/platform/platform-impl/src/com/intellij/ui/popup/AbstractPopup.java?L475-493
    private Window getParentWindow(Component component) {
        Window window = null;
        Component parent = UIUtil.findUltimateParent(component == null ? WindowManagerEx.getInstanceEx().getFocusedComponent(project) : component);
        if (parent instanceof Window) {
            window = (Window) parent;
        }
        if (window == null) {
            window = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusedWindow();
        }
        return window;
    }

    @Override
    public void dispose() {
        if (popup != null) {
            popup.getWindow().dispose();
        }

        mainPanel.dispose();
    }


    // We manually emit an action defined by the material UI theme to hide the overlay it opens whenever a popover is
    // created. This third-party plugin does not work with our approach of keeping the popover alive and thus, when the
    // Sourcegraph popover is closed, their custom overlay stays active.
    //
    //   - https://github.com/sourcegraph/sourcegraph/issues/36479
    //   - https://github.com/mallowigi/material-theme-issues/issues/179
    private void hideMaterialUiOverlay() {
        AnAction materialAction = ActionManager.getInstance().getAction("MTToggleOverlaysAction");
        if (materialAction != null) {
            try {
                DataContext dataContext = DataManager.getInstance().getDataContextFromFocusAsync().blockingGet(10);
                if (dataContext != null) {
                    materialAction.actionPerformed(
                        new AnActionEvent(
                            null,
                            dataContext,
                            ActionPlaces.UNKNOWN,
                            new Presentation(),
                            ActionManager.getInstance(),
                            0)
                    );
                }
            } catch (Exception ignored) {
            }
        }
    }
}
