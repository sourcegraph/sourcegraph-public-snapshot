package com.sourcegraph.cody.chat;

import static com.sourcegraph.cody.chat.ChatUIConstants.TEXT_MARGIN;

import com.intellij.ui.components.JBTextField;
import com.intellij.util.ui.JBInsets;
import com.sourcegraph.cody.api.Speaker;
import com.sourcegraph.cody.context.ContextFile;
import com.sourcegraph.cody.context.ContextMessage;
import com.sourcegraph.cody.ui.AccordionSection;
import java.awt.*;
import java.util.List;
import java.util.Objects;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;
import javax.swing.border.EmptyBorder;
import org.jetbrains.annotations.NotNull;

public class ContextFilesMessage extends MessagePanel {
  public ContextFilesMessage(@NotNull List<ContextMessage> contextMessages) {
    super(Speaker.ASSISTANT, ChatUIConstants.ASSISTANT_MESSAGE_GRADIENT_WIDTH);
    this.setLayout(new BorderLayout());

    JBInsets margin =
        JBInsets.create(new Insets(TEXT_MARGIN, TEXT_MARGIN, TEXT_MARGIN, TEXT_MARGIN));
    List<String> contextFileNames =
        contextMessages.stream()
            .map(ContextMessage::getFile)
            .filter(Objects::nonNull)
            .map(ContextFile::getFileName)
            .collect(Collectors.toList());

    AccordionSection accordionSection =
        new AccordionSection("Read " + contextFileNames.size() + " files");
    accordionSection.setOpaque(false);
    accordionSection.setBorder(new EmptyBorder(margin));
    AtomicInteger fileIndex = new AtomicInteger(0);
    contextFileNames.forEach(
        fileName -> {
          JBTextField textField = new JBTextField(fileName);
          textField.setOpaque(false);
          textField.setMargin(margin);
          textField.setEditable(false);
          accordionSection.getContentPanel().add(textField, fileIndex.getAndIncrement());
        });
    this.add(accordionSection, BorderLayout.CENTER);
  }
}
