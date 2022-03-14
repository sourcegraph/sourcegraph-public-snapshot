import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.ide.CopyPasteManager;

import java.awt.datatransfer.StringSelection;

public class Copy extends FileAction {

    @Override
    void handleFileUri(String uri) {
        // Remove utm tags for sharing
        String shortenURI = uri.replaceAll("(&utm_product_name=)(.*)", "");
        // Copy file uri to clipboard
        CopyPasteManager.getInstance().setContents(new StringSelection(shortenURI));

        // Display bubble
        Notification notification = new Notification("Sourcegraph", "Sourcegraph",
                "File URL copied to clipboard."+shortenURI, NotificationType.INFORMATION);
//        Editor.getProject
//        NotificationGroupManager.getInstance().getNotificationGroup("Sourcegraph")
//                .createNotification("File URL copied to clipboard."+shortenURI, NotificationType.INFORMATION)
//                .notify(this.);
        Notifications.Bus.notify(notification);
    }
}
