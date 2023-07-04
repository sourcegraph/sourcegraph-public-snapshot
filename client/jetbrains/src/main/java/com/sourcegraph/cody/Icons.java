package com.sourcegraph.cody;

import com.intellij.openapi.util.IconLoader;
import javax.swing.*;

public interface Icons {
  Icon CodyLogo = IconLoader.getIcon("/icons/codyLogo.svg", Icons.class);

  interface Repository {
    Icon Indexed = IconLoader.getIcon("/icons/repositoryIndexed.svg", Icons.class);
    Icon Missing = IconLoader.getIcon("/icons/repositoryMissing.svg", Icons.class);
  }
}
