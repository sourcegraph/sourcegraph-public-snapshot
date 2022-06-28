package com.sourcegraph.config;

import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Properties;

public class UserLevelConfig {
    @Nullable
    public static String getDefaultBranchName() {
        Properties properties = readProperties();
        return properties.getProperty("defaultBranch", null);
    }

    @Nullable
    public static String getRemoteUrlReplacements() {
        Properties properties = readProperties();
        return properties.getProperty("remoteUrlReplacements", null);
    }

    @NotNull
    public static String getSourcegraphUrl() {
        Properties properties = readProperties();
        String url = properties.getProperty("url", "https://sourcegraph.com/");
        return url.endsWith("/") ? url : url + "/";
    }

    // readProps returns the first properties file it's able to parse from the following paths:
    //   $HOME/.sourcegraph-jetbrains.properties
    //   $HOME/sourcegraph-jetbrains.properties
    @NotNull
    private static Properties readProperties() {
        Path[] candidatePaths = {
            Paths.get(System.getProperty("user.home"), ".sourcegraph-jetbrains.properties"),
            Paths.get(System.getProperty("user.home"), "sourcegraph-jetbrains.properties"),
        };

        for (Path path : candidatePaths) {
            try {
                return readPropertiesFile(path.toFile());
            } catch (IOException e) {
                // no-op
            }
        }
        // No files found/readable
        return new Properties();
    }

    @NotNull
    private static Properties readPropertiesFile(File file) throws IOException {
        Properties properties = new Properties();

        try (InputStream input = new FileInputStream(file)) {
            properties.load(input);
        }

        return properties;
    }
}
