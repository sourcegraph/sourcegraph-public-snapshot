import com.google.common.base.Strings;
import java.io.UnsupportedEncodingException;
import java.net.URI;
import java.net.URLEncoder;

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
      remoteURL = repoInfo.remoteURL.replace(".git", "").replaceFirst(":", "/").replace("git@", "https://");;
    }
    URI remote = URI.create(remoteURL);
    String path = remote.getPath();

    StringBuilder builder = new StringBuilder();
    try {
      builder.append(sourcegraphBase);
      builder.append(String.format("/%s%s", remote.getHost(), path));
      builder.append(String.format("/-/commit/%s", revisionNumber));
      builder.append(String.format("?editor=%s", URLEncoder.encode("JetBrains", "UTF-8")));
      builder.append(String.format("&version=%s", URLEncoder.encode(Util.VERSION, "UTF-8")));
      builder.append(String.format("&utm_product_name=%s", URLEncoder.encode(productName, "UTF-8")));
      builder.append(String.format("&utm_product_version=%s", URLEncoder.encode(productVersion, "UTF-8")));
    } catch (UnsupportedEncodingException e) {
      e.printStackTrace();
    }

    return URI.create(builder.toString());
  }

}
