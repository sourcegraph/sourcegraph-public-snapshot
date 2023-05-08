package com.sourcegraph.cody.chat;

import javax.swing.*;
import java.awt.*;

public class ChatBubble extends JPanel {
    private int radius;

    public ChatBubble() {
        super();
    }

    public int getRadius() {
        return radius;
    }

    public void setRadius(int radius) {
        this.radius = radius;
    }

    @Override
    protected void paintComponent(Graphics g) {
        final Graphics2D g2d = (Graphics2D) g;
        g2d.setColor(getBackground());
        g2d.fillRoundRect(0, 0, this.getWidth() - 1, this.getHeight() - 1, this.radius, this.radius);

        // print debug info
        System.out.println("chatBubble sizes:");
        System.out.println("  width: " + this.getWidth());
        System.out.println("  height: " + this.getHeight());
        System.out.println("  radius: " + this.radius);
    }
}
