package com.sourcegraph.common;

import com.intellij.ide.BrowserUtil;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.config.ConfigUtil;
import org.jetbrains.annotations.NotNull;

import java.awt.*;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;

public class BrowserOpener {
    public static void openRelativeUrlInBrowser(@NotNull Project project, @NotNull String relativeUrl) {
        openInBrowser(project, ConfigUtil.getSourcegraphUrl(project) + "/" + relativeUrl);
    }

    public static void openInBrowser(@NotNull Project project, @NotNull String absoluteUrl) {
        try {
            openInBrowser(project, new URI(absoluteUrl));
        } catch (URISyntaxException e) {
            Logger logger = Logger.getInstance(BrowserOpener.class);
            logger.warn("Error while creating URL from \"" + absoluteUrl + "\": " + e.getMessage());
        }
    }

    public static void openInBrowser(@NotNull Project project, @NotNull URI uri) {
        try {
            BrowserUtil.browse(uri);
        } catch (Exception e) {
            try {
                Desktop.getDesktop().browse(uri);
            } catch (IOException | UnsupportedOperationException e2) {
                BrowserErrorNotification.show(project, uri);
            }
        }
    }
}
