import static org.junit.jupiter.api.Assertions.*;

import com.sourcegraph.cody.autocomplete.CodyAutocompleteManager;
import difflib.Patch;
import org.junit.jupiter.api.Test;

public class AutocompleteDiffTest {
  @Test
  public void minimalDiff() {
    Patch<String> patch = CodyAutocompleteManager.diff("println()", "println(arrays());");
    // NOTE(olafurpg): ideally, we should get the delta size to 1. Myer's diff seems to emit
    // unnecessary deltas that we might be able to merge to reduce the number of displayed inlay
    // hints.
    assertEquals(2, patch.getDeltas().size());
  }
}
