package com.sourcegraph.cody.chat;

import com.intellij.ui.JBColor;
import com.intellij.util.ui.JBInsets;
import com.intellij.util.ui.UIUtil;
import com.sourcegraph.cody.completions.Speaker;

import javax.swing.*;
import javax.swing.border.EmptyBorder;
import java.awt.*;

public class ChatBubble extends JPanel {
    private int radius;

    public ChatBubble() {
        super();
    }

    public ChatBubble(int radius, Color background, ChatMessage message) {
        super();

        boolean isHuman = message.getSpeaker() == Speaker.HUMAN;
        this.radius = radius;
        this.setBackground(background);
        this.setBorder(new EmptyBorder(new JBInsets(10, 10, 10, 10)));
        JTextArea bubbleTextArea = new JTextArea(message.getDisplayText());
        bubbleTextArea.setFont(UIUtil.getLabelFont());
        bubbleTextArea.setLineWrap(true);
        bubbleTextArea.setWrapStyleWord(true);
        bubbleTextArea.setBackground(background);
        bubbleTextArea.setForeground(isHuman ? JBColor.WHITE : JBColor.BLACK);
        bubbleTextArea.setComponentOrientation(isHuman ? ComponentOrientation.RIGHT_TO_LEFT : ComponentOrientation.LEFT_TO_RIGHT);
        this.add(bubbleTextArea, BorderLayout.CENTER);
    }

    public void setRadius(int radius) {
        this.radius = radius;
    }

    @Override
    protected void paintComponent(Graphics g) {
        final Graphics2D g2d = (Graphics2D) g;
        g2d.setColor(getBackground());
        g2d.fillRoundRect(0, 0, this.getWidth() - 1, this.getHeight() - 1, this.radius, this.radius);
    }
}
