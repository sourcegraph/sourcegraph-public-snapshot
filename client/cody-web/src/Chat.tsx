import React, { useRef } from 'react'

import { noop } from 'lodash'

import { Chat as ChatUI, ChatUISubmitButtonProps, ChatUITextAreaProps } from '@sourcegraph/cody-ui/src/Chat'
import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'
import { SubmitSvg } from '@sourcegraph/cody-ui/src/utils/icons'

import styles from './Chat.module.css'

export const Chat: React.FunctionComponent<
    Omit<
        React.ComponentPropsWithoutRef<typeof ChatUI>,
        | 'textAreaComponent'
        | 'submitButtonComponent'
        | 'fileLinkComponent'
        | 'messageBeingEdited'
        | 'setMessageBeingEdited'
    >
> = ({
    messageInProgress,
    transcript,
    contextStatus,
    formInput,
    setFormInput,
    inputHistory,
    setInputHistory,
    onSubmit,
}) => (
    <ChatUI
        messageBeingEdited={false}
        setMessageBeingEdited={noop}
        messageInProgress={messageInProgress}
        transcript={transcript}
        contextStatus={contextStatus}
        formInput={formInput}
        setFormInput={setFormInput}
        inputHistory={inputHistory}
        setInputHistory={setInputHistory}
        onSubmit={onSubmit}
        textAreaComponent={TextArea}
        submitButtonComponent={SubmitButton}
        fileLinkComponent={FileLink}
        afterTips={CODY_TERMS_MARKDOWN}
        transcriptItemClassName={styles.transcriptItem}
        humanTranscriptItemClassName={styles.humanTranscriptItem}
        transcriptItemParticipantClassName={styles.transcriptItemParticipant}
        transcriptActionClassName={styles.transcriptAction}
        inputRowClassName={styles.inputRow}
        chatInputClassName={styles.chatInput}
        isCodyEnabled={true}
    />
)

const TextArea: React.FunctionComponent<ChatUITextAreaProps> = ({
    className,
    rows,
    autoFocus,
    value,
    required,
    onInput,
    onKeyDown,
}) => {
    const textAreaRef = useRef<HTMLTextAreaElement>(null)

    const handleKeyDown = (event: React.KeyboardEvent<HTMLElement>): void => {
        if (onKeyDown) {
            onKeyDown(event, textAreaRef.current?.selectionStart ?? null)
        }
    }

    return (
        <textarea
            ref={textAreaRef}
            className={className}
            rows={rows}
            value={value}
            autoFocus={autoFocus}
            required={required}
            onInput={onInput}
            onKeyDown={handleKeyDown}
        />
    )
}

const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = ({ className, disabled, onClick }) => (
    <button className={className} type="submit" disabled={disabled} onClick={onClick}>
        <SubmitSvg />
    </button>
)

const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>
