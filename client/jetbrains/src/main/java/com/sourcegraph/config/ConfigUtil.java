package com.sourcegraph.config;

import com.intellij.openapi.project.Project;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Objects;
import java.util.Properties;

public class ConfigUtil {
    public static String getDefaultBranchNameSetting(Project project) {
        String defaultBranch = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getDefaultBranch();
        if (defaultBranch == null || defaultBranch.length() == 0) {
            Properties props = readProps();
            defaultBranch = props.getProperty("defaultBranch", null);
        }
        return defaultBranch;
    }

    public static String getRemoteUrlReplacements(Project project) {
        String replacements = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getRemoteUrlReplacements();
        if (replacements == null || replacements.length() == 0) {
            Properties props = readProps();
            replacements = props.getProperty("remoteUrlReplacements", null);
        }
        return replacements;
    }

    public static String getSourcegraphUrl(Project project) {
        String url = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getUrl();
        if (url == null || url.length() == 0) {
            Properties props = readProps();
            url = props.getProperty("url", "https://sourcegraph.com/");
        }
        return url.endsWith("/") ? url : url + "/";
    }

    public static String getVersion() {
        return "v1.2.2";
    }

    // readProps returns the first properties file it's able to parse from the following paths:
    //   $HOME/.sourcegraph-jetbrains.properties
    //   $HOME/sourcegraph-jetbrains.properties
    private static Properties readProps() {
        Path[] candidatePaths = {
            Paths.get(System.getProperty("user.home"), ".sourcegraph-jetbrains.properties"),
            Paths.get(System.getProperty("user.home"), "sourcegraph-jetbrains.properties"),
        };

        for (Path path : candidatePaths) {
            try {
                return readPropsFile(path.toFile());
            } catch (IOException e) {
                // no-op
            }
        }
        // No files found/readable
        return new Properties();
    }

    private static Properties readPropsFile(File file) throws IOException {
        Properties props = new Properties();

        try (InputStream input = new FileInputStream(file)) {
            props.load(input);
        }

        return props;
    }
}
