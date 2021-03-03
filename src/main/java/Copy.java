import com.intellij.notification.Notification;
import com.intellij.notification.NotificationType;
import com.intellij.notification.Notifications;
import com.intellij.openapi.ide.CopyPasteManager;

import java.awt.datatransfer.StringSelection;

public class Copy extends FileAction {

    @Override
    void handleFileUri(String uri) {
        // Copy file uri to clipboard
        CopyPasteManager.getInstance().setContents(new StringSelection(uri));

        // Display bubble
        Notification notification = new Notification("Sourcegraph", "Sourcegraph",
                "File URL copied to clipboard.", NotificationType.INFORMATION);
        Notifications.Bus.notify(notification);
    }
}