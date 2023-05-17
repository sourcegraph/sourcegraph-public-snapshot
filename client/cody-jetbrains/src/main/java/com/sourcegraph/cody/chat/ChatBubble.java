package com.sourcegraph.cody.chat;

import com.intellij.ui.JBColor;
import com.intellij.util.ui.JBInsets;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.completions.Speaker;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import javax.swing.border.EmptyBorder;
import java.awt.*;

public class ChatBubble extends JPanel {
    private final int radius;

    public ChatBubble(int radius, @NotNull ChatMessage message) {
        super();

        boolean isHuman = message.getSpeaker() == Speaker.HUMAN;
        JBColor background = isHuman ? JBColor.BLUE : JBColor.GRAY;
        this.radius = radius;
        this.setBackground(background);
        this.setLayout(new BorderLayout());
        this.setBorder(new EmptyBorder(new JBInsets(10, 10, 10, 10)));
        JTextArea textArea = new JTextArea(message.getDisplayText());
        textArea.setFont(UIUtil.getLabelFont());
        textArea.setLineWrap(true);
        textArea.setWrapStyleWord(true);
        textArea.setBackground(background);
        textArea.setForeground(isHuman ? JBColor.WHITE : JBColor.BLACK);
        textArea.setComponentOrientation(isHuman ? ComponentOrientation.RIGHT_TO_LEFT : ComponentOrientation.LEFT_TO_RIGHT);
        this.add(textArea, BorderLayout.CENTER);
    }

    public void updateText(@NotNull String newText) {
        JTextArea textArea = (JTextArea) this.getComponent(0);
        textArea.setText(newText);
    }

    @Override
    protected void paintComponent(@NotNull Graphics g) {
        final Graphics2D g2d = (Graphics2D) g;
        g2d.setColor(getBackground());
        g2d.fillRoundRect(0, 0, this.getWidth() - 1, this.getHeight() - 1, this.radius, this.radius);
    }
}
