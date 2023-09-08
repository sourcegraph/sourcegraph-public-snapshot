package com.sourcegraph.website;

import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.ide.CopyPasteManager;
import com.intellij.openapi.project.Project;
import java.awt.datatransfer.StringSelection;
import org.jetbrains.annotations.NotNull;

public class CopyAction extends FileActionBase {
  @Override
  protected void handleFileUri(@NotNull Project project, @NotNull String uri) {
    // Remove utm tags for sharing
    String urlWithoutUtm = uri.replaceAll("(&utm_product_name=)(.*)", "");

    // Copy file uri to clipboard
    CopyPasteManager.getInstance().setContents(new StringSelection(urlWithoutUtm));

    // Display notification
    Notification notification =
        new Notification(
            "Sourcegraph: URL sharing",
            "Sourcegraph",
            "File URL copied to clipboard: " + urlWithoutUtm,
            NotificationType.INFORMATION);
    Notifications.Bus.notify(notification);
  }
}
