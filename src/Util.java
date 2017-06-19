import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.VisualPosition;
import org.jetbrains.annotations.NotNull;
import sun.plugin.dom.exception.InvalidStateException;
import com.intellij.openapi.editor.LogicalPosition;

import java.io.*;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.Properties;

public class Util {
    public static String VERSION = "v1.1.1";

    // gitRemotes returns the names of all git remotes, e.g. ["origin", "foobar"]
    public static String[] gitRemotes(String repoDir) throws IOException {
        return exec("git remote", repoDir).split("[\\r\\n]+");
    }

    // gitRemoteURL returns the remote URL for the given remote name.
    // e.g. "origin" -> "git@github.com:foo/bar"
    public static String gitRemoteURL(String repoDir, String remoteName) throws IOException {
        return exec("git remote get-url " + remoteName, repoDir).trim();
    }

    // gitDefaultRemoteURL returns the remote URL of the first Git remote
    // found. An exception is thrown if there is not one.
    public static String gitDefaultRemoteURL(String repoDir) throws Exception {
        String[] remotes = gitRemotes(repoDir);
        if (remotes.length == 0) {
            throw new Exception("no configured git remotes");
        }
        if (remotes.length > 1) {
            Logger.getInstance(Util.class).info("using first git remote: " + remotes[0]);
        }
        return gitRemoteURL(repoDir, remotes[0]);
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

    public static String sourcegraphURL() {
        Properties props = readProps();
        String url = props.getProperty("url", "https://sourcegraph.com");
        if (!url.endsWith("/")) {
            return url + "/";
        }
        return url;
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
            remoteURL = gitDefaultRemoteURL(repoRoot);
            branch = gitBranch(repoRoot);
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
