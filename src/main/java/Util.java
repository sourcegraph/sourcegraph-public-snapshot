import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.project.Project;

import java.io.*;
import java.util.Properties;

public class Util {
    public static String VERSION = "v1.2.0";

    // gitRemoteURL returns the remote URL for the given remote name.
    // e.g. "origin" -> "git@github.com:foo/bar"
    public static String gitRemoteURL(String repoDir, String remoteName) throws Exception {
        String s = exec("git remote get-url " + remoteName, repoDir).trim();
        if (s.isEmpty()) {
            throw new Exception("no such remote");
        }
        return s;
    }

    // configuredGitRemoteURL returns the URL of the "sourcegraph" remote, if
    // configured, or else the URL of the "origin" remote. An exception is
    // thrown if neither exists.
    public static String configuredGitRemoteURL(String repoDir) throws Exception {
        try {
            return gitRemoteURL(repoDir, "sourcegraph");
        } catch (Exception err) {
            try {
                return gitRemoteURL(repoDir, "origin");
            } catch (Exception err2) {
                throw new Exception("no configured git remote \"sourcegraph\" or \"origin\"");
            }
        }
    }

    // gitRootDir returns the repository root directory for any directory
    // within the repository.
    public static String gitRootDir(String repoDir) throws IOException {
        return exec("git rev-parse --show-toplevel", repoDir).trim();
    }

    // gitBranch returns either the current branch name of the repository OR in
    // all other cases (e.g. detached HEAD state), it returns "HEAD".
    public static String gitBranch(String repoDir) throws IOException {
        return exec("git rev-parse --abbrev-ref HEAD", repoDir).trim();
    }

    // verify that provided branch exists on remote
    public static boolean isRemoteBranch(String branch, String repoDir) throws IOException {
        return exec("git show-branch remotes/origin/" + branch, repoDir).length() > 0;
    }

    public static String sourcegraphURL(Project project) {
        String url = Config.getInstance(project).getUrl();
        if (url == null || url.length() == 0) {
            Properties props = readProps();
            url = props.getProperty("url", "https://sourcegraph.com/");
        }
        return url.endsWith("/") ? url : url + "/";
    }

    // readProps tries to read the $HOME/sourcegraph-jetbrains.properties file.
    private static Properties readProps() {
        Properties props = new Properties();
        InputStream input = null;

        String path = System.getProperty("user.home") + File.separator + "sourcegraph-jetbrains.properties";
        try{
            input = new FileInputStream(path);
            props.load(input);
        } catch (IOException e) {
            // no-op
        } finally {
            if (input != null) {
                try{
                    input.close();
                } catch (IOException e) {
                    // no-op
                }
            }
        }
        return props;
    }

    // repoInfo returns the Sourcegraph repository URI, and the file path
    // relative to the repository root. If the repository URI cannot be
    // determined, a RepoInfo with empty strings is returned.
    public static RepoInfo repoInfo(String fileName) {
        String fileRel = "";
        String remoteURL = "";
        String branch = "";
        try{
            // Determine repository root directory.
            String fileDir = fileName.substring(0, fileName.lastIndexOf("/"));
            String repoRoot = gitRootDir(fileDir);

            // Determine file path, relative to repository root.
            fileRel = fileName.substring(repoRoot.length()+1);
            remoteURL = configuredGitRemoteURL(repoRoot);
            branch = gitBranch(repoRoot);

            // If on a branch that does not exist on the remote, use "master" instead.
            if (!isRemoteBranch(branch, repoRoot)) {
                branch = "master";
            }
        } catch (Exception err) {
            Logger.getInstance(Util.class).info(err);
            err.printStackTrace();
        }
        return new RepoInfo(fileRel, remoteURL, branch);
    }

    // exec executes the given command in the specified directory and returns
    // its stdout. Any stderr output is logged.
    public static String exec(String cmd, String dir) throws IOException {
        Logger.getInstance(Util.class).debug("exec cmd='" + cmd + "' dir="+dir);

        // Create the process.
        Process p = Runtime.getRuntime().exec(cmd, null, new File(dir));
        BufferedReader stdout = new BufferedReader(new InputStreamReader(p.getInputStream()));
        BufferedReader stderr = new BufferedReader(new InputStreamReader(p.getErrorStream()));

        // Log any stderr ouput.
        Logger logger = Logger.getInstance(Util.class);
        String s;
        while ((s = stderr.readLine()) != null) {
            logger.debug(s);
        }

        String out = new String();
        for (String l; (l = stdout.readLine()) != null; out += l + "\n");
        return out;
    }
}
