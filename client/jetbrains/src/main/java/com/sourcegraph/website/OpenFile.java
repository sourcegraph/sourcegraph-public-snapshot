package com.sourcegraph.website;

import com.sourcegraph.common.BrowserErrorNotification;
import org.jetbrains.annotations.NotNull;

import java.awt.*;
import java.io.IOException;
import java.net.URI;

public class OpenFile extends FileActionBase {
    @Override
    protected void handleFileUri(@NotNull String url) {
        URI uri = URI.create(url);

        try {
            Desktop.getDesktop().browse(uri);
        } catch (IOException | UnsupportedOperationException e) {
            BrowserErrorNotification.show(uri);
        }
    }
}
