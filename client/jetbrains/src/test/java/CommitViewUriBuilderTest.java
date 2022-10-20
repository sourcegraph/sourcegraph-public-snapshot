import com.sourcegraph.git.CommitViewUriBuilder;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EmptySource;

import java.net.URI;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

public class CommitViewUriBuilderTest {

  @Test
  public void testBuild_AllValid() {
    CommitViewUriBuilder builder = new CommitViewUriBuilder();

    String remoteUrl = "https://github.com/sourcegraph/sourcegraph-jetbrains.git";

    URI got = builder.build("https://www.sourcegraph.com",
        "1fa8d5d6286c24924b55c15ed4d1a0b85c44b4d5",
        remoteUrl,
        "intellij",
        "1.1");

    String want = "https://www.sourcegraph.com/github.com/sourcegraph/sourcegraph-jetbrains.git/-/commit/1fa8d5d6286c24924b55c15ed4d1a0b85ccab4d5?editor=JetBrains&version=v1.2.2&utm_product_name=intellij&utm_product_version=1.1";
    assertEquals(want, got.toString());
  }

  @ParameterizedTest
  @EmptySource
  public void testBuild_MissingRevision(String revision) {
      CommitViewUriBuilder builder = new CommitViewUriBuilder();
      String remoteUrl = "https://github.com/sourcegraph/sourcegraph-jetbrains.git";

    assertThrows(RuntimeException.class, () -> builder.build("https://www.sourcegraph.com",
        revision,
        remoteUrl,
        "intellij",
        "1.1"));
  }

  @ParameterizedTest
  @EmptySource
  public void testBuild_MissingBaseUri(String baseUri) {
      CommitViewUriBuilder builder = new CommitViewUriBuilder();
      String remoteUrl = "https://github.com/sourcegraph/sourcegraph-jetbrains.git";

    assertThrows(RuntimeException.class, () -> builder.build(baseUri,
        "1fa8d5d6286c24924b55c15ed4d1a0b85c44b4d5",
        remoteUrl,
        "intellij",
        "1.1"));
  }
  @Test
  public void testBuild_MissingRemoteUrl() {
      CommitViewUriBuilder builder = new CommitViewUriBuilder();
      String remoteUrl = "";

      assertThrows(RuntimeException.class, () -> builder.build("https://www.sourcegraph.com",
          "1fa8d5d6286c24924b55c15ed4d1a0b85c44b4d5",
          remoteUrl,
          "intellij",
          "1.1"));
  }
}
