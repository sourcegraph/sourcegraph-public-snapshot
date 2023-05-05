import React, { useCallback, useEffect, useRef, useState } from 'react'

import { VSCodeButton, VSCodeTextArea } from '@vscode/webview-ui-toolkit/react'
import classNames from 'classnames'

import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { escapeCodyMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import {
    Chat as ChatUI,
    ChatUISubmitButtonProps,
    ChatUISuggestionButtonProps,
    ChatUITextAreaProps,
    EditButtonProps,
    FeedbackButtonsProps,
} from '@sourcegraph/cody-ui/src/Chat'
import { SubmitSvg } from '@sourcegraph/cody-ui/src/utils/icons'

import { FileLink } from './FileLink'
import { VSCodeWrapper } from './utils/VSCodeApi'

import styles from './Chat.module.css'

interface ChatboxProps {
    messageInProgress: ChatMessage | null
    messageBeingEdited: boolean
    setMessageBeingEdited: (input: boolean) => void
    transcript: ChatMessage[]
    contextStatus: ChatContextStatus | null
    formInput: string
    setFormInput: (input: string) => void
    inputHistory: string[]
    setInputHistory: (history: string[]) => void
    vscodeAPI: VSCodeWrapper
    suggestions?: string[]
    setSuggestions?: (suggestions: undefined | string[]) => void
}

export const Chat: React.FunctionComponent<React.PropsWithChildren<ChatboxProps>> = ({
    messageInProgress,
    messageBeingEdited,
    setMessageBeingEdited,
    transcript,
    contextStatus,
    formInput,
    setFormInput,
    inputHistory,
    setInputHistory,
    vscodeAPI,
    suggestions,
    setSuggestions,
}) => {
    const onSubmit = useCallback(
        (text: string, submitType: 'user' | 'suggestion') => {
            vscodeAPI.postMessage({ command: 'submit', text: escapeCodyMarkdown(text), submitType })
        },
        [vscodeAPI]
    )

    const onEditBtnClick = useCallback(
        (text: string) => {
            vscodeAPI.postMessage({ command: 'edit', text })
        },
        [vscodeAPI]
    )

    const onFeedbackBtnClick = useCallback(
        (text: string) => {
            vscodeAPI.postMessage({ command: 'event', event: 'feedback', value: text })
        },
        [vscodeAPI]
    )

    const onCopyBtnClick = useCallback(
        (text: string) => {
            vscodeAPI.postMessage({ command: 'event', event: 'click', value: text })
        },
        [vscodeAPI]
    )

    return (
        <ChatUI
            messageInProgress={messageInProgress}
            messageBeingEdited={messageBeingEdited}
            setMessageBeingEdited={setMessageBeingEdited}
            transcript={transcript}
            contextStatus={contextStatus}
            formInput={formInput}
            setFormInput={setFormInput}
            inputHistory={inputHistory}
            setInputHistory={setInputHistory}
            onSubmit={onSubmit}
            textAreaComponent={TextArea}
            submitButtonComponent={SubmitButton}
            suggestionButtonComponent={SuggestionButton}
            fileLinkComponent={FileLink}
            className={styles.innerContainer}
            codeBlocksCopyButtonClassName={styles.codeBlocksCopyButton}
            transcriptItemClassName={styles.transcriptItem}
            humanTranscriptItemClassName={styles.humanTranscriptItem}
            transcriptItemParticipantClassName={styles.transcriptItemParticipant}
            transcriptActionClassName={styles.transcriptAction}
            inputRowClassName={styles.inputRow}
            chatInputContextClassName={styles.chatInputContext}
            chatInputClassName={styles.chatInputClassName}
            EditButtonContainer={EditButton}
            editButtonOnSubmit={onEditBtnClick}
            FeedbackButtonsContainer={FeedbackButtons}
            feedbackButtonsOnSubmit={onFeedbackBtnClick}
            copyButtonOnSubmit={onCopyBtnClick}
            suggestions={suggestions}
            setSuggestions={setSuggestions}
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

const SuggestionButton: React.FunctionComponent<ChatUISuggestionButtonProps> = ({ suggestion, onClick }) => (
    <VSCodeButton className={styles.suggestionButton} appearance="secondary" type="button" onClick={onClick}>
        <i className="codicon codicon-sparkle" slot="start">
            {/* Fallback emoji because this icon is a new addition and doesn't seem to work for me? */}✨
        </i>{' '}
        {suggestion}
    </VSCodeButton>
)

const EditButton: React.FunctionComponent<EditButtonProps> = ({
    className,
    messageBeingEdited,
    setMessageBeingEdited,
}) => (
    <div className={className}>
        <VSCodeButton
            className={classNames(styles.editButton)}
            appearance="secondary"
            type="button"
            onClick={() => setMessageBeingEdited(!messageBeingEdited)}
        >
            <i className={messageBeingEdited ? 'codicon codicon-close' : 'codicon codicon-edit'} />
        </VSCodeButton>
    </div>
)

const FeedbackButtons: React.FunctionComponent<FeedbackButtonsProps> = ({ className, feedbackButtonsOnSubmit }) => {
    const [feedbackSubmitted, setFeedbackSubmitted] = useState(false)

    const onFeedbackBtnSubmit = useCallback(
        (text: string) => {
            feedbackButtonsOnSubmit(text)
            setFeedbackSubmitted(true)
        },
        [feedbackButtonsOnSubmit]
    )

    if (feedbackSubmitted) {
        return (
            <div className={className}>
                <VSCodeButton className={classNames(styles.submitButton)} title="Feedback submitted." disabled={true}>
                    <i className="codicon codicon-check" />
                </VSCodeButton>
            </div>
        )
    }

    return (
        <div className={classNames(styles.feedbackButtons, className)}>
            <VSCodeButton
                className={classNames(styles.submitButton)}
                appearance="icon"
                type="button"
                onClick={() => onFeedbackBtnSubmit('thumbsUp')}
            >
                <i className="codicon codicon-thumbsup" />
            </VSCodeButton>
            <VSCodeButton
                className={classNames(styles.submitButton)}
                appearance="icon"
                type="button"
                onClick={() => onFeedbackBtnSubmit('thumbsDown')}
            >
                <i className="codicon codicon-thumbsdown" />
            </VSCodeButton>
        </div>
    )
}
