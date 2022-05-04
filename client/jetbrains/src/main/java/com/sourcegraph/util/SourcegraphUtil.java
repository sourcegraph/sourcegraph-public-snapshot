package com.sourcegraph.util;

import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;
import com.sourcegraph.project.RepoInfo;
import com.sourcegraph.project.SourcegraphConfig;

import java.io.*;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Objects;
import java.util.Properties;

public class SourcegraphUtil {
    public static final String VERSION = "v1.2.2";

    public static String sourcegraphURL(Project project) {
        String url = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getUrl();
        if (url == null || url.length() == 0) {
            Properties props = readProps();
            url = props.getProperty("url", "https://sourcegraph.com/");
        }
        return url.endsWith("/") ? url : url + "/";
    }

    // get defaultBranch configuration option
    public static String getDefaultBranchNameSetting(Project project) {
        String defaultBranch = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getDefaultBranch();
        if (defaultBranch == null || defaultBranch.length() == 0) {
            Properties props = readProps();
            defaultBranch = props.getProperty("defaultBranch", null);
        }
        return defaultBranch;
    }

    // get remoteUrlReplacements configuration option
    public static String setRemoteUrlReplacements(Project project) {
        String replacements = Objects.requireNonNull(SourcegraphConfig.getInstance(project)).getRemoteUrlReplacements();
        if (replacements == null || replacements.length() == 0) {
            Properties props = readProps();
            replacements = props.getProperty("remoteUrlReplacements", null);
        }
        return replacements;
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

    // repoInfo returns the Sourcegraph repository URI, and the file path
    // relative to the repository root. If the repository URI cannot be
    // determined, a RepoInfo with empty strings is returned.
    public static RepoInfo repoInfo(String filePath, Project project) {
        String relativePath = "";
        String remoteUrl = "";
        String branchName = "";
        try {
            String defaultBranchNameSetting = SourcegraphUtil.getDefaultBranchNameSetting(project);
            String repoRootPath = GitUtil.getRepoRootPath(filePath);

            // Determine file path, relative to repository root.
            relativePath = filePath.substring(repoRootPath.length() + 1);
            remoteUrl = GitUtil.getConfiguredRemoteUrl(repoRootPath);
            branchName = defaultBranchNameSetting != null ? defaultBranchNameSetting : GitUtil.getCurrentBranchName(repoRootPath);

            // If on a branch that does not exist on the remote and no defaultBranch is configured
            // use "master" instead.
            // This allows users to check out a branch that does not exist in origin remote by setting defaultBranch
            if (!GitUtil.doesRemoteBranchExist(branchName, repoRootPath) && defaultBranchNameSetting == null) {
                branchName = "master"; // TODO:
            }

            // replace remoteURL if config option is not null
            String r = SourcegraphUtil.setRemoteUrlReplacements(project);
            if (r != null) {
                String[] replacements = r.trim().split("\\s*,\\s*");
                // Check if the entered values are pairs
                for (int i = 0; i < replacements.length && replacements.length % 2 == 0; i += 2) {
                    remoteUrl = remoteUrl.replace(replacements[i], replacements[i + 1]);
                }
            }
        } catch (Exception err) {
            Logger.getInstance(SourcegraphUtil.class).info(err);
            err.printStackTrace();
        }
        return new RepoInfo(relativePath, remoteUrl, branchName);
    }

    // exec executes the given command in the specified directory and returns
    // its stdout. Any stderr output is logged.
    public static String exec(String command, String directoryPath) throws IOException {
        Logger.getInstance(SourcegraphUtil.class).debug("exec cmd='" + command + "' dir=" + directoryPath);

        // Create the process.
        Process p = Runtime.getRuntime().exec(command, null, new File(directoryPath));
        BufferedReader stdout = new BufferedReader(new InputStreamReader(p.getInputStream()));
        BufferedReader stderr = new BufferedReader(new InputStreamReader(p.getErrorStream()));

        // Log any stderr output.
        Logger logger = Logger.getInstance(SourcegraphUtil.class);
        String s;
        while ((s = stderr.readLine()) != null) {
            logger.debug(s);
        }

        String out = "";
        //noinspection StatementWithEmptyBody
        for (String l; (l = stdout.readLine()) != null; out += l + "\n") ;
        return out;
    }
}
