package com.sourcegraph.website;

import com.intellij.openapi.diagnostic.Logger;
import org.jetbrains.annotations.NotNull;

import java.awt.*;
import java.io.IOException;
import java.net.URI;

public class OpenFile extends FileActionBase {
    @Override
    protected void handleFileUri(@NotNull String uri) {
        Logger logger = Logger.getInstance(this.getClass());

        // Open the URL in the browser.
        try {
            Desktop.getDesktop().browse(URI.create(uri));
        } catch (IOException err) {
            logger.debug("Failed to open browser.", err);
            err.printStackTrace();
        }
    }
}
