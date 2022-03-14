import com.google.common.base.Strings;
import java.net.URI;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;

public class CommitViewUriBuilder {

  public URI build(String sourcegraphBase, String revisionNumber, RepoInfo repoInfo, String productName, String productVersion) {
    if (Strings.isNullOrEmpty(sourcegraphBase)) {
      throw new RuntimeException("Missing sourcegraph URI for commit uri.");
    } else if (Strings.isNullOrEmpty(revisionNumber)) {
      throw new RuntimeException("Missing revision number for commit uri.");
    } else if (repoInfo == null || Strings.isNullOrEmpty(repoInfo.remoteURL)) {
      throw new RuntimeException("Missing remote URL for commit uri.");
    }

    // this is pretty hacky but to try to build the repo string we will just try to naively parse the git remote uri. Worst case scenario this 404s
    String remoteURL = repoInfo.remoteURL;
    if(remoteURL.startsWith("git")){
      remoteURL = repoInfo.remoteURL.replace(".git", "").replaceFirst(":", "/").replace("git@", "https://");
    }
    URI remote = URI.create(remoteURL);
    String path = remote.getPath();

      String url = sourcegraphBase +
              String.format("/%s%s", remote.getHost(), path) +
              String.format("/-/commit/%s", revisionNumber) +
              String.format("?editor=%s", URLEncoder.encode("JetBrains", StandardCharsets.UTF_8)) +
              String.format("&version=%s", URLEncoder.encode(Util.VERSION, StandardCharsets.UTF_8)) +
              String.format("&utm_product_name=%s", URLEncoder.encode(productName, StandardCharsets.UTF_8)) +
              String.format("&utm_product_version=%s", URLEncoder.encode(productVersion, StandardCharsets.UTF_8));

    return URI.create(url);
  }

}
