package com.sourcegraph.cody.ui;

import com.intellij.ui.JBColor;
import com.intellij.util.ui.JBUI;
import java.awt.*;

public class Colors {
  public static final Color CYAN = new JBColor(new Color(0, 203, 236), new Color(0, 203, 236));
  public static final Color PURPLE = new JBColor(new Color(161, 18, 255), new Color(161, 18, 255));
  public static final Color ORANGE = new JBColor(new Color(255, 85, 67), new Color(255, 85, 67));
  public static final Color SECONDARY_LINK_COLOR =
      new JBColor(JBUI.CurrentTheme.Link.Foreground.ENABLED, new Color(168, 173, 189));
}
