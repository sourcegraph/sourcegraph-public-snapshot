import React, { useCallback, useEffect, useRef } from 'react'

import { VSCodeButton, VSCodeTextArea } from '@vscode/webview-ui-toolkit/react'
import classNames from 'classnames'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { Chat as ChatUI, ChatUISubmitButtonProps, ChatUITextAreaProps } from '@sourcegraph/cody-ui/src/Chat'
import { SubmitSvg, ResetIcon } from '@sourcegraph/cody-ui/src/utils/icons'

import { FileLink } from './FileLink'
import { vscodeAPI } from './utils/VSCodeApi'

import styles from './Chat.module.css'

interface ChatboxProps {
    messageInProgress: ChatMessage | null
    transcript: ChatMessage[]
    formInput: string
    setFormInput: (input: string) => void
    inputHistory: string[]
    setInputHistory: (history: string[]) => void
}

const TIPS_RECOMMENDATIONS: JSX.Element[] = [
    <>Visit the `Recipes` tab for special actions like Generate a unit test or Summarize recent code changes.</>,
    <>
        Use the <ResetIcon /> button in the upper right to reset the chat when you want to start a new line of thought.
        Cody does not remember anything outside the current chat.
    </>,
]

export const Chat: React.FunctionComponent<React.PropsWithChildren<ChatboxProps>> = ({
    messageInProgress,
    transcript,
    formInput,
    setFormInput,
    inputHistory,
    setInputHistory,
}) => {
    const onSubmit = useCallback((text: string) => {
        vscodeAPI.postMessage({ command: 'submit', text })
    }, [])

    return (
        <ChatUI
            messageInProgress={messageInProgress}
            transcript={transcript}
            formInput={formInput}
            setFormInput={setFormInput}
            inputHistory={inputHistory}
            setInputHistory={setInputHistory}
            onSubmit={onSubmit}
            textAreaComponent={TextArea}
            submitButtonComponent={SubmitButton}
            fileLinkComponent={FileLink}
            tipsRecommendations={TIPS_RECOMMENDATIONS}
            className={styles.innerContainer}
            codeBlocksCopyButtonClassName={styles.codeBlocksCopyButton}
            transcriptItemClassName={styles.transcriptItem}
            humanTranscriptItemClassName={styles.humanTranscriptItem}
            transcriptItemParticipantClassName={styles.transcriptItemParticipant}
            inputRowClassName={styles.inputRow}
            chatInputClassName={styles.chatInputClassName}
        />
    )
}

const TextArea: React.FunctionComponent<ChatUITextAreaProps> = ({
    className,
    rows,
    autoFocus,
    value,
    required,
    onInput,
    onKeyDown,
}) => {
    // Focus the textarea when the webview gains focus (unless there is text selected). This makes
    // it so that the user can immediately start typing to Cody after invoking `Cody: Focus on Chat
    // View` with the keyboard.
    const inputRef = useRef<HTMLElement>(null)
    useEffect(() => {
        const handleFocus = (): void => {
            if (document.getSelection()?.isCollapsed) {
                inputRef.current?.focus()
            }
        }
        window.addEventListener('focus', handleFocus)
        return () => {
            window.removeEventListener('focus', handleFocus)
        }
    }, [])

    // <VSCodeTextArea autofocus> does not work, so implement autofocus ourselves.
    useEffect(() => {
        if (autoFocus) {
            inputRef.current?.focus()
        }
    }, [autoFocus])

    return (
        <VSCodeTextArea
            className={classNames(styles.chatInput, className)}
            rows={rows}
            ref={
                // VSCodeTextArea has a very complex type.
                //
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                inputRef as any
            }
            value={value}
            autofocus={autoFocus}
            required={required}
            onInput={e => onInput(e as React.FormEvent<HTMLTextAreaElement>)}
            onKeyDown={onKeyDown}
        />
    )
}

const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = ({ className, disabled, onClick }) => (
    <VSCodeButton
        className={classNames(styles.submitButton, className)}
        appearance="icon"
        type="button"
        disabled={disabled}
        onClick={onClick}
    >
        <SubmitSvg />
    </VSCodeButton>
)
