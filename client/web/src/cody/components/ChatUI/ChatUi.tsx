import React, { useCallback, useEffect, useRef, useState, useMemo } from 'react'

import { mdiClose, mdiSend, mdiArrowDown, mdiPencil, mdiThumbUp, mdiThumbDown, mdiCheck } from '@mdi/js'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import {
    Chat,
    ChatUISubmitButtonProps,
    ChatUITextAreaProps,
    EditButtonProps,
    FeedbackButtonsProps,
} from '@sourcegraph/cody-ui/src/Chat'
import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'
import { Button, Icon, TextArea, Link, Tooltip, Alert, Text, H2 } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../tracking/eventLogger'
import { CodyPageIcon } from '../../chat/CodyPageIcon'
import { useCodySidebar } from '../../sidebar/Provider'
import { CodyChatStore } from '../../useCodyChat'
import { ScopeSelector } from '../ScopeSelector'

import styles from './ChatUi.module.scss'

export const SCROLL_THRESHOLD = 100

const onFeedbackSubmit = (feedback: string): void => eventLogger.log(`web:cody:feedbackSubmit:${feedback}`)

interface IChatUIProps {
    codyChatStore: CodyChatStore
}

export const ChatUI: React.FC<IChatUIProps> = ({ codyChatStore }): JSX.Element => {
    const {
        submitMessage,
        editMessage,
        messageInProgress,
        chatMessages,
        transcript,
        transcriptHistory,
        loaded,
        isCodyEnabled,
        scope,
        setScope,
        toggleIncludeInferredRepository,
        toggleIncludeInferredFile,
    } = codyChatStore

    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>(() =>
        transcriptHistory
            .flatMap(entry => entry.interactions)
            .sort((entryA, entryB) => +new Date(entryA.timestamp) - +new Date(entryB.timestamp))
            .filter(interaction => interaction.humanMessage.displayText !== undefined)
            .map(interaction => interaction.humanMessage.displayText!)
    )
    const [messageBeingEdited, setMessageBeingEdited] = useState<boolean>(false)

    const onSubmit = useCallback((text: string) => submitMessage(text), [submitMessage])
    const onEdit = useCallback((text: string) => editMessage(text), [editMessage])

    const scopeSelectorProps = useMemo(
        () => ({ scope, setScope, toggleIncludeInferredRepository, toggleIncludeInferredFile }),
        [scope, setScope, toggleIncludeInferredRepository, toggleIncludeInferredFile]
    )

    if (!loaded) {
        return <></>
    }

    return (
        <>
            <Chat
                key={transcript?.id}
                messageInProgress={messageInProgress}
                messageBeingEdited={messageBeingEdited}
                setMessageBeingEdited={setMessageBeingEdited}
                transcript={chatMessages}
                formInput={formInput}
                setFormInput={setFormInput}
                inputHistory={inputHistory}
                setInputHistory={setInputHistory}
                onSubmit={onSubmit}
                submitButtonComponent={SubmitButton}
                fileLinkComponent={FileLink}
                className={styles.container}
                afterTips={CODY_TERMS_MARKDOWN}
                transcriptItemClassName={styles.transcriptItem}
                humanTranscriptItemClassName={styles.humanTranscriptItem}
                transcriptItemParticipantClassName="text-muted"
                inputRowClassName={styles.inputRow}
                chatInputClassName={styles.chatInput}
                EditButtonContainer={EditButton}
                editButtonOnSubmit={onEdit}
                textAreaComponent={AutoResizableTextArea}
                codeBlocksCopyButtonClassName={styles.codeBlocksCopyButton}
                transcriptActionClassName={styles.transcriptAction}
                FeedbackButtonsContainer={FeedbackButtons}
                feedbackButtonsOnSubmit={onFeedbackSubmit}
                needsEmailVerification={isCodyEnabled.needsEmailVerification}
                needsEmailVerificationNotice={NeedsEmailVerificationNotice}
                contextStatusComponent={ScopeSelector}
                contextStatusComponentProps={scopeSelectorProps}
            />
        </>
    )
}

export const ScrollDownButton = React.memo(function ScrollDownButtonContent({
    onClick,
}: {
    onClick: () => void
}): JSX.Element {
    return (
        <div className={styles.scrollButtonWrapper}>
            <Button className={styles.scrollButton} onClick={onClick}>
                <Icon aria-label="Scroll down" svgPath={mdiArrowDown} />
            </Button>
        </div>
    )
})

export const EditButton: React.FunctionComponent<EditButtonProps> = React.memo(function EditButtonContent({
    className,
    messageBeingEdited,
    setMessageBeingEdited,
}) {
    return (
        <div className={className}>
            <button
                className={classNames(className, styles.editButton)}
                type="button"
                onClick={() => setMessageBeingEdited(!messageBeingEdited)}
            >
                {messageBeingEdited ? (
                    <Icon aria-label="Close" svgPath={mdiClose} />
                ) : (
                    <Icon aria-label="Edit" svgPath={mdiPencil} />
                )}
            </button>
        </div>
    )
})

const FeedbackButtons: React.FunctionComponent<FeedbackButtonsProps> = React.memo(function FeedbackButtonsContent({
    feedbackButtonsOnSubmit,
}) {
    const [feedbackSubmitted, setFeedbackSubmitted] = useState(false)

    const onFeedbackBtnSubmit = useCallback(
        (text: string) => {
            feedbackButtonsOnSubmit(text)
            setFeedbackSubmitted(true)
        },
        [feedbackButtonsOnSubmit]
    )

    return (
        <div className={classNames('d-flex', styles.feedbackButtonsWrapper)}>
            {feedbackSubmitted ? (
                <Button title="Feedback submitted." disabled={true} className="ml-1 p-1">
                    <Icon aria-label="Feedback submitted" svgPath={mdiCheck} />
                </Button>
            ) : (
                <>
                    <Button
                        title="Thumbs up"
                        className="ml-1 p-1"
                        type="button"
                        onClick={() => onFeedbackBtnSubmit('positive')}
                    >
                        <Icon aria-label="Thumbs up" svgPath={mdiThumbUp} />
                    </Button>
                    <Button
                        title="Thumbs up"
                        className="ml-1 p-1"
                        type="button"
                        onClick={() => onFeedbackBtnSubmit('negative')}
                    >
                        <Icon aria-label="Thumbs down" svgPath={mdiThumbDown} />
                    </Button>
                </>
            )}
        </div>
    )
})

export const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = React.memo(function SubmitButtonContent({
    className,
    disabled,
    onClick,
}) {
    return (
        <button
            className={classNames(className, styles.submitButton)}
            type="submit"
            disabled={disabled}
            onClick={onClick}
        >
            <Icon aria-label="Submit" svgPath={mdiSend} />
        </button>
    )
})

export const FileLink: React.FunctionComponent<FileLinkProps> = React.memo(function FileLinkContent({
    path,
    repoName,
    revision,
}) {
    return repoName ? (
        <Link to={`/${repoName}${revision ? `@${revision}` : ''}/-/blob/${path}`}>{path}</Link>
    ) : (
        <>{path}</>
    )
})

interface AutoResizableTextAreaProps extends ChatUITextAreaProps {}

export const AutoResizableTextArea: React.FC<AutoResizableTextAreaProps> = React.memo(
    function AutoResizableTextAreaContent({ value, onInput, onKeyDown, className, disabled = false }) {
        const { inputNeedsFocus, setFocusProvided, isCodyEnabled } = useCodySidebar() || {
            inputNeedsFocus: false,
            setFocusProvided: () => null,
        }
        const textAreaRef = useRef<HTMLTextAreaElement>(null)
        const { width = 0 } = useResizeObserver({ ref: textAreaRef })

        const adjustTextAreaHeight = useCallback((): void => {
            if (textAreaRef.current) {
                textAreaRef.current.style.height = '0px'
                const scrollHeight = textAreaRef.current.scrollHeight
                textAreaRef.current.style.height = `${scrollHeight}px`

                // Hide scroll if the textArea isn't overflowing.
                textAreaRef.current.style.overflowY = scrollHeight < 200 ? 'hidden' : 'auto'
            }
        }, [])

        const handleChange = (): void => {
            adjustTextAreaHeight()
        }

        useEffect(() => {
            if (inputNeedsFocus && textAreaRef.current) {
                textAreaRef.current.focus()
                setFocusProvided()
            }
        }, [inputNeedsFocus, setFocusProvided])

        useEffect(() => {
            adjustTextAreaHeight()
        }, [adjustTextAreaHeight, value, width])

        const handleKeyDown = (event: React.KeyboardEvent<HTMLElement>): void => {
            if (onKeyDown) {
                onKeyDown(event, textAreaRef.current?.selectionStart ?? null)
            }
        }

        return (
            <Tooltip content={isCodyEnabled.needsEmailVerification ? 'Verify your email to use Cody.' : ''}>
                <TextArea
                    ref={textAreaRef}
                    className={className}
                    value={value}
                    onChange={handleChange}
                    rows={1}
                    autoFocus={false}
                    required={true}
                    onKeyDown={handleKeyDown}
                    onInput={onInput}
                    disabled={disabled}
                />
            </Tooltip>
        )
    }
)

const NeedsEmailVerificationNotice: React.FunctionComponent = React.memo(
    function NeedsEmailVerificationNoticeContent() {
        return (
            <div className="p-3">
                <H2 className={classNames('d-flex gap-1 align-items-center mb-3', styles.codyMessageHeader)}>
                    <CodyPageIcon /> Cody
                </H2>
                <Alert variant="warning">
                    <Text className="mb-0">Verify email</Text>
                    <Text className="mb-0">
                        Using Cody requires a verified email.{' '}
                        <Link to={`${window.context.currentUser?.settingsURL}/emails`} target="_blank" rel="noreferrer">
                            Resend email verification
                        </Link>
                        .
                    </Text>
                </Alert>
            </div>
        )
    }
)
