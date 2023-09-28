package com.sourcegraph.cody;

import com.intellij.openapi.util.IconLoader;
import com.intellij.ui.AnimatedIcon;
import javax.swing.*;

public interface Icons {
  Icon CodyLogo = IconLoader.getIcon("/icons/codyLogo.svg", Icons.class);
  Icon HiImCody = IconLoader.getIcon("/icons/hiImCodyLogo.svg", Icons.class);

  interface Repository {
    Icon Indexed = IconLoader.getIcon("/icons/repositoryIndexed.svg", Icons.class);
    Icon NoEmbedding = IconLoader.getIcon("/icons/repositoryNoEmbedding.svg", Icons.class);
    Icon Missing = IconLoader.getIcon("/icons/repositoryMissing.svg", Icons.class);
  }

  interface Actions {
    Icon Hide = IconLoader.getIcon("/icons/actions/hide.svg", Icons.class);
  }

  interface StatusBar {
    Icon CompletionInProgress = new AnimatedIcon.Default();
    Icon CodyAvailable = IconLoader.getIcon("/icons/codyLogoMonochromatic.svg", Icons.class);
    Icon CodyAutocompleteDisabled =
        IconLoader.getIcon("/icons/codyLogoMonochromaticMuted.svg", Icons.class);

    Icon CodyAutocompleteUnavailable =
        IconLoader.getIcon("/icons/codyLogoMonochromaticUnavailable.svg", Icons.class);
  }

  interface Onboarding {
    Icon Autocomplete = IconLoader.getIcon("/icons/onboarding/autocomplete.svg", Icons.class);
    Icon Chat = IconLoader.getIcon("/icons/onboarding/chat.svg", Icons.class);
    Icon Commands = IconLoader.getIcon("/icons/onboarding/commands.svg", Icons.class);
  }
}
