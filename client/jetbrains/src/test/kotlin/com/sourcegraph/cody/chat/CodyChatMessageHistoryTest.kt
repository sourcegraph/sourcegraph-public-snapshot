package com.sourcegraph.cody.chat

import com.intellij.ui.components.JBTextArea
import com.sourcegraph.config.ConfigUtil
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class CodyChatMessageHistoryTest {
    private lateinit var history: CodyChatMessageHistory;

    @BeforeEach
    fun setup() {
        history = CodyChatMessageHistory(10);
    }
    @Test
    fun `messageSent adds message to latest position in history`() {
        val textArea = JBTextArea()
        textArea.text = "test"
        history.messageSent(textArea)

        history.popUpperMessage(textArea)
        assertThat(textArea.text).isEqualTo("test")
    }

    @Test
    fun `popUpperMessage pops message from upper stack`() {
        val textArea = JBTextArea()
        textArea.text = "test 1"
        history.messageSent(textArea)
        textArea.text = "test 2"
        history.messageSent(textArea)

        history.popUpperMessage(textArea)
        assertThat(textArea.text).isEqualTo("test 2")

        history.popUpperMessage(textArea)
        assertThat(textArea.text).isEqualTo("test 1")
    }

    @Test
    fun `popLowerMessages pop message from lower stack`() {
        val textArea = JBTextArea()
        textArea.text = "test 1"
        history.messageSent(textArea)
        textArea.text = "test 2"
        history.messageSent(textArea)

        history.popUpperMessage(textArea)
        assertThat(textArea.text).isEqualTo("test 2")

        history.popLowerMessage(textArea)
        assertThat(textArea.text).isEmpty()
    }

    @Test
    fun `popUpperMessage stops at last message`() {
        val textArea = JBTextArea()
        textArea.text = "test 1"
        history.messageSent(textArea)

        history.popUpperMessage(textArea)
        assertThat(textArea.text).isEqualTo("test 1")
        history.popUpperMessage(textArea)
        assertThat(textArea.text).isEqualTo("test 1")
    }

    @Test
    fun `popLowerMessage stops at empty string`() {
        val textArea = JBTextArea()
        textArea.text = "test 1"
        history.messageSent(textArea)
        history.popUpperMessage(textArea)
        history.popLowerMessage(textArea)

        history.popLowerMessage(textArea)
        assertThat(textArea.text).isEmpty()
        history.popLowerMessage(textArea)
        assertThat(textArea.text).isEmpty()
    }

    @Test
    fun `popUpperMessage overrides oldest message when capacity exceeded`() {
        val textArea = JBTextArea()
        val testString = "test"
        for (i in 1..10){
            textArea.text = testString.plus(i)
            history.messageSent(textArea)
        }
        textArea.text = "last test"
        history.messageSent(textArea)

        history.popUpperMessage(textArea)
        assertThat(textArea.text).isEqualTo("last test")

        for (i in 1..11){
            history.popUpperMessage(textArea)
        }
        assertThat(textArea.text).isEqualTo("test2")

    }
}
