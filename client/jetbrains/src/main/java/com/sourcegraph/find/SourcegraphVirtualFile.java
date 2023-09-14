package com.sourcegraph.find;

import com.intellij.testFramework.LightVirtualFile;
import com.intellij.util.LocalTimeCounter;
import org.jetbrains.annotations.NotNull;

public class SourcegraphVirtualFile extends LightVirtualFile {
  private final String repoUrl;
  private final String commit;
  private final String path;

  public SourcegraphVirtualFile(
      @NotNull String name,
      @NotNull CharSequence content,
      String repoUrl,
      String commit,
      String path) {
    super(name, null, content, LocalTimeCounter.currentTime());
    this.repoUrl = repoUrl;
    this.commit = commit;
    this.path = path;
  }

  public String getRepoUrl() {
    return repoUrl;
  }

  public String getCommit() {
    return commit;
  }

  @NotNull
  public String getRelativePath() {
    return path;
  }

  @NotNull
  @Override
  public String getPath() {
    return repoUrl + " > " + path;
  }
}
