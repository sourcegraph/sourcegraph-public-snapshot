package com.sourcegraph.find;

import com.intellij.util.ui.JBEmptyBorder;
import com.intellij.util.ui.components.BorderLayoutPanel;
import com.sourcegraph.Icons;
import com.sourcegraph.config.GoToPluginSettingsButtonFactory;
import java.awt.*;
import javax.swing.*;

public class HeaderPanel extends BorderLayoutPanel {
  public HeaderPanel() {
    super();
    setBorder(new JBEmptyBorder(5, 5, 2, 5));

    JPanel title = new JPanel(new FlowLayout(FlowLayout.LEFT, 0, 0));
    title.setBorder(new JBEmptyBorder(2, 0, 0, 0));
    title.add(new JLabel("Find with Sourcegraph", Icons.SourcegraphLogo, SwingConstants.LEFT));

    JPanel buttons = new JPanel(new FlowLayout(FlowLayout.RIGHT, 0, 0));
    buttons.add(GoToPluginSettingsButtonFactory.createGoToPluginSettingsButton());

    add(title, BorderLayout.WEST);
    add(buttons, BorderLayout.EAST);
  }
}
