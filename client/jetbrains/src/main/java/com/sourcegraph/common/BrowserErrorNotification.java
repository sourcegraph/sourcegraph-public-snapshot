package com.sourcegraph.common;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.DumbAwareAction;
import com.sourcegraph.Icons;
import org.jetbrains.annotations.NotNull;

import java.awt.*;
import java.awt.datatransfer.Clipboard;
import java.awt.datatransfer.StringSelection;
import java.net.URI;

public class BrowserErrorNotification {
    public static void show(URI uri) {
        Notification notification = new Notification("Sourcegraph errors", "Sourcegraph",
            "Opening external browser is not supported. You can still copy the URL to your clipboard and open it manually.", NotificationType.WARNING);
        AnAction copyUrlAction = new DumbAwareAction("Copy URL") {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                String url = uri.toString();
                StringSelection selection = new StringSelection(url);
                Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();
                clipboard.setContents(selection, selection);
                notification.expire();
            }
        };
        AnAction dismissAction = new DumbAwareAction("Dismiss") {
            @Override
            public void actionPerformed(@NotNull AnActionEvent anActionEvent) {
                notification.expire();
            }
        };
        notification.setIcon(Icons.SourcegraphLogo);
        notification.addAction(copyUrlAction);
        notification.addAction(dismissAction);
        Notifications.Bus.notify(notification);
    }
}
