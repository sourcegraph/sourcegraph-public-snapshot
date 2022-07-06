package com.sourcegraph;

import com.intellij.openapi.util.IconLoader;

import javax.swing.*;

public interface Icons {
    Icon Logo = IconLoader.getIcon("/icons/sourcegraph-logo.png", Icons.class);
    Icon Account = IconLoader.getIcon("/icons/account.svg", Icons.class);
    Icon GearPlain = IconLoader.getIcon("/icons/gearPlain.svg", Icons.class);
}
