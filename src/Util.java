import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.editor.VisualPosition;
import org.jetbrains.annotations.NotNull;
import sun.plugin.dom.exception.InvalidStateException;
import com.intellij.openapi.editor.LogicalPosition;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.MalformedURLException;
import java.net.URL;

public class Util {
    public static String VERSION = "v1.0.2";

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

    // removePrefixes removes any of the given prefixes from the input string
    // `s`. Only one prefix is removed.
    public static String removePrefixes(String s, String[] prefixes) {
        for(int i = 0; i < prefixes.length; i++) {
            if (s.startsWith(prefixes[i])) {
                return s.substring(prefixes[i].length());
            }
        }
        return s;
    }

    // replaceLastOccurrence returns `s` with the last occurrence of `a`
    // replaced by `b`.
    public static String replaceLastOccurrence(String s, String a, String b) {
        int k = s.lastIndexOf(a);
        if(k == -1) {
            return s;
        }
        return s.substring(0, k) + b + s.substring(k+1);
    }

    // repoFromRemoteURL returns the repository name from the remote URL. An
    // exception is raised if it cannot be determined. Supported formats are:
    //
    // 	optional("ssh://" OR "git://" OR "https://" OR "https://")
    // 	+ optional("username") + optional(":password") + optional("@")
    // 	+ "github.com"
    // 	+ "/" OR ":"
    // 	+ "<organization>" + "/" + "<username>"
    //
    public static String repoFromRemoteURL(String remoteURL) throws MalformedURLException, Exception {
        // Normalize all URL schemes into 'http://' just for parsing purposes.
        // We don't actually care about the scheme itself.
        String r = removePrefixes(remoteURL, new String[]{"ssh://", "git://", "https://", "http://"});

        // Normalize github.com:foo/bar -> github.com/foo/bar -- Note we only
        // do the last occurrence as it may be included earlier in the case of
        // 'foo:bar@github.com'
        r = replaceLastOccurrence(r, ":", "/");

        URL u = new URL("http://" + r);
        if(!u.getHost().endsWith("github.com")) { // Note: using endsWith because getHost may have 'username:password@' prefix.
            throw new Exception("repository remote is not github.com " + remoteURL);
        }
        return "github.com" + u.getPath();
    }

    public static String sourcegraphURL() {
        String url = "https://sourcegraph.com"; // TODO: Make this user configurable!
        if (!url.endsWith("/")) {
            return url + "/";
        }
        return url;
    }

    public static String lineHash(LogicalPosition start, LogicalPosition end) {
        if(start == null || end == null) {
            return "";
        }
        return "#L" + Integer.toString(start.line+1) + ":" + Integer.toString(start.column+1) + "-" + Integer.toString(end.line+1) + ":" + Integer.toString(end.column+1);
    }

    public static String branchStr(String branch) {
        if (branch.equals("HEAD")) {
            return ""; // Detached HEAD state
        }
        if (branch.equals("master")) {
            // Assume master is the default branch, for now.
            return "";
        }
        return "@" + branch;
    }

    // repoInfo returns the Sourcegraph repository URI, and the file path
    // relative to the repository root. If the repository URI cannot be
    // determined, an exception is thrown.
    public static RepoInfo repoInfo(String fileName) throws IOException, Exception, MalformedURLException {
        // Determine repository root directory.
        String fileDir = fileName.substring(0, fileName.lastIndexOf("/"));
        String repoRoot = gitRootDir(fileDir);

        // Determine file path, relative to repository root.
        String fileRel = fileName.substring(repoRoot.length()+1);
        String repo = repoFromRemoteURL(gitDefaultRemoteURL(repoRoot));
        String branch = gitBranch(repoRoot);
        return new RepoInfo(fileRel, repo, branch);
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
