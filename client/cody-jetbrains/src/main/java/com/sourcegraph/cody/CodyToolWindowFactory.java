package com.sourcegraph.cody;

import com.intellij.openapi.project.DumbAware;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.wm.ToolWindow;
import com.intellij.openapi.wm.ToolWindowFactory;
import com.intellij.ui.content.Content;
import com.intellij.ui.content.ContentFactory;
import org.apache.commons.lang.StringUtils;
import org.jetbrains.annotations.NotNull;

import javax.swing.*;
import java.awt.*;
import java.util.Calendar;
import java.util.Objects;

public class CodyToolWindowFactory implements ToolWindowFactory, DumbAware {
    @Override
    public boolean isApplicable(@NotNull Project project) {
        return ToolWindowFactory.super.isApplicable(project);
    }

    @Override
    public void createToolWindowContent(@NotNull Project project, @NotNull ToolWindow toolWindow) {
        CodyToolWindowContent toolWindowContent = new CodyToolWindowContent(toolWindow);
        Content content = ContentFactory.SERVICE.getInstance().createContent(toolWindowContent.getContentPanel(), "", false);
        toolWindow.getContentManager().addContent(content);
    }

    private static class CodyToolWindowContent {
        private final JPanel contentPanel = new JPanel();
        private final JLabel currentDate = new JLabel();
        private final JLabel timeZone = new JLabel();
        private final JLabel currentTime = new JLabel();

        public CodyToolWindowContent(@NotNull ToolWindow toolWindow) {
            contentPanel.setLayout(new BorderLayout(0, 20));
            contentPanel.setBorder(BorderFactory.createEmptyBorder(40, 0, 0, 0));
            contentPanel.add(createCalendarPanel(), BorderLayout.PAGE_START);
            contentPanel.add(createControlsPanel(toolWindow), BorderLayout.CENTER);
            updateCurrentDateTime();
        }

        @NotNull
        private JPanel createCalendarPanel() {
            JPanel calendarPanel = new JPanel();
            calendarPanel.add(currentDate);
            calendarPanel.add(timeZone);
            calendarPanel.add(currentTime);
            return calendarPanel;
        }

        private void setIconLabel(JLabel label, String imagePath) {
            label.setIcon(new ImageIcon(Objects.requireNonNull(getClass().getResource(imagePath))));
        }

        @NotNull
        private JPanel createControlsPanel(ToolWindow toolWindow) {
            JPanel controlsPanel = new JPanel();
            JButton refreshDateAndTimeButton = new JButton("Refresh");
            refreshDateAndTimeButton.addActionListener(e -> updateCurrentDateTime());
            controlsPanel.add(refreshDateAndTimeButton);
            JButton hideToolWindowButton = new JButton("Hide");
            hideToolWindowButton.addActionListener(e -> toolWindow.hide(null));
            controlsPanel.add(hideToolWindowButton);
            return controlsPanel;
        }

        private void updateCurrentDateTime() {
            Calendar calendar = Calendar.getInstance();
            currentDate.setText(getCurrentDate(calendar));
            timeZone.setText(getTimeZone(calendar));
            currentTime.setText(getCurrentTime(calendar));
        }

        private String getCurrentDate(Calendar calendar) {
            return calendar.get(Calendar.DAY_OF_MONTH) + "/"
                + (calendar.get(Calendar.MONTH) + 1) + "/"
                + calendar.get(Calendar.YEAR);
        }

        private String getTimeZone(Calendar calendar) {
            long gmtOffset = calendar.get(Calendar.ZONE_OFFSET); // offset from GMT in milliseconds
            String gmtOffsetString = String.valueOf(gmtOffset / 3600000);
            return (gmtOffset > 0) ? "GMT + " + gmtOffsetString : "GMT - " + gmtOffsetString;
        }

        private String getCurrentTime(Calendar calendar) {
            return getFormattedValue(calendar, Calendar.HOUR_OF_DAY) + ":" + getFormattedValue(calendar, Calendar.MINUTE);
        }

        private String getFormattedValue(Calendar calendar, int calendarField) {
            int value = calendar.get(calendarField);
            return StringUtils.leftPad(Integer.toString(value), 2, "0");
        }

        public JPanel getContentPanel() {
            return contentPanel;
        }
    }
}
