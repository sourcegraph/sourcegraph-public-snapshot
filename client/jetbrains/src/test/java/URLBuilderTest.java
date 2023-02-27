import com.sourcegraph.website.URLBuilder;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EmptySource;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

public class URLBuilderTest {

    @Test
    public void testBuildCommitUrl_AllValid() {
        String remoteUrl = "https://github.com/sourcegraph/sourcegraph-jetbrains.git";

        String got = URLBuilder.buildCommitUrl("https://www.sourcegraph.com",
            "1fa8d5d6286c24924b55c15ed4d1a0b85c44b4d5",
            remoteUrl,
            "intellij",
            "1.1");

        String want = "https://www.sourcegraph.com/github.com/sourcegraph/sourcegraph-jetbrains.git/-/commit/1fa8d5d6286c24924b55c15ed4d1a0b85ccab4d5?editor=JetBrains&version=v1.2.2&utm_product_name=intellij&utm_product_version=1.1";
        assertEquals(want, got);
    }

    @ParameterizedTest
    @EmptySource
    public void testBuildCommitUrl_MissingRevision(String revision) {
        String remoteUrl = "https://github.com/sourcegraph/sourcegraph-jetbrains.git";

        assertThrows(RuntimeException.class, () -> URLBuilder.buildCommitUrl("https://www.sourcegraph.com",
            revision,
            remoteUrl,
            "intellij",
            "1.1"));
    }

    @ParameterizedTest
    @EmptySource
    public void testBuildCommitUrl_MissingBaseUri(String baseUri) {
        String remoteUrl = "https://github.com/sourcegraph/sourcegraph-jetbrains.git";

        assertThrows(RuntimeException.class, () -> URLBuilder.buildCommitUrl(baseUri,
            "1fa8d5d6286c24924b55c15ed4d1a0b85c44b4d5",
            remoteUrl,
            "intellij",
            "1.1"));
    }

    @Test
    public void testBuildCommitUrl_MissingRemoteUrl() {
        String remoteUrl = "";

        assertThrows(RuntimeException.class, () -> URLBuilder.buildCommitUrl("https://www.sourcegraph.com",
            "1fa8d5d6286c24924b55c15ed4d1a0b85c44b4d5",
            remoteUrl,
            "intellij",
            "1.1"));
    }
}
