import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.net.URI;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EmptySource;
import org.junit.jupiter.params.provider.NullSource;

public class CommitViewUriBuilderTest {

  @Test
  public void testBuild_AllValid() {
    CommitViewUriBuilder builder = new CommitViewUriBuilder();

    RepoInfo repoInfo = new RepoInfo("", "https://github.com/sourcegraph/sourcegraph-jetbrains.git", "main");

    URI got = builder.build("https://www.sourcegraph.com",
        "1fa8d5d6286c24924b55c15ed4d1a0b85ccab4d5",
        repoInfo,
        "intellij",
        "1.1");

    String want = "https://www.sourcegraph.com/github.com/sourcegraph/sourcegraph-jetbrains.git/-/commit/1fa8d5d6286c24924b55c15ed4d1a0b85ccab4d5?editor=JetBrains&version=v1.2.2&utm_product_name=intellij&utm_product_version=1.1";
    assertEquals(want, got.toString());
  }

  @ParameterizedTest
  @NullSource
  @EmptySource
  public void testBuild_MissingRevision(String revision) {
    CommitViewUriBuilder builder = new CommitViewUriBuilder();
    RepoInfo repoInfo = new RepoInfo("", "https://github.com/sourcegraph/sourcegraph-jetbrains.git", "main");

    assertThrows(RuntimeException.class, () -> builder.build("https://www.sourcegraph.com",
        revision,
        repoInfo,
        "intellij",
        "1.1"));
  }

  @ParameterizedTest
  @NullSource
  @EmptySource
  public void testBuild_MissingBaseUri(String baseUri) {
    CommitViewUriBuilder builder = new CommitViewUriBuilder();
    RepoInfo repoInfo = new RepoInfo("", "https://github.com/sourcegraph/sourcegraph-jetbrains.git", "main");

    assertThrows(RuntimeException.class, () -> builder.build(baseUri,
        "1fa8d5d6286c24924b55c15ed4d1a0b85ccab4d5",
        repoInfo,
        "intellij",
        "1.1"));
  }
  @Test
  public void testBuild_MissingRemoteUrl() {
    CommitViewUriBuilder builder = new CommitViewUriBuilder();
    RepoInfo repoInfo = new RepoInfo("", "", "main");

    assertThrows(RuntimeException.class, () -> builder.build("https://www.sourcegraph.com",
        "1fa8d5d6286c24924b55c15ed4d1a0b85ccab4d5",
        repoInfo,
        "intellij",
        "1.1"));

    assertThrows(RuntimeException.class, () -> builder.build("https://www.sourcegraph.com",
        "1fa8d5d6286c24924b55c15ed4d1a0b85ccab4d5",
        null,
        "intellij",
        "1.1"));
  }
}