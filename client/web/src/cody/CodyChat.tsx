import { useCallback, useEffect, useRef, useState } from 'react'

import { mdiClose, mdiSend, mdiArrowDown, mdiReload, mdiPencil, mdiHistory } from '@mdi/js'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { Chat, ChatUISubmitButtonProps, ChatUITextAreaProps, EditButtonProps } from '@sourcegraph/cody-ui/src/Chat'
import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'
import { CodyLogo } from '@sourcegraph/cody-ui/src/icons/CodyLogo'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'
import { Button, Icon, TextArea, Tooltip } from '@sourcegraph/wildcard'

import { useChatStoreState } from '../stores/codyChat'

import { ChatHistory } from './ChatHistory'

import styles from './CodyChat.module.scss'

export const SCROLL_THRESHOLD = 100

interface CodyChatProps {
    onClose: () => void
}

export const CodyChat = ({ onClose }: CodyChatProps): JSX.Element => {
    const {
        reset,
        submitMessage,
        editMessage,
        messageInProgress,
        transcript,
        getChatContext,
        transcriptHistory,
        loadTranscriptFromHistory,
        clearHistory,
    } = useChatStoreState()

    const codySidebarRef = useRef<HTMLDivElement>(null)
    const [formInput, setFormInput] = useState('')
    const [showHistory, setShowHistory] = useState(false)
    const [inputHistory, setInputHistory] = useState<string[] | []>([])
    const [shouldScrollToBottom, setShouldScrollToBottom] = useState(true)
    const [showScrollDownButton, setShowScrollDownButton] = useState(false)
    const [messageBeingEdited, setMessageBeingEdited] = useState<boolean>(false)

    const handleScroll = useCallback(() => {
        if (codySidebarRef.current) {
            const { scrollHeight, scrollTop, clientHeight } = codySidebarRef.current
            const scrollOffset = scrollHeight - scrollTop - clientHeight
            setShouldScrollToBottom(scrollOffset <= SCROLL_THRESHOLD)
            setShowScrollDownButton(scrollOffset > SCROLL_THRESHOLD)
        }
    }, [codySidebarRef])

    const scrollToBottom = (behavior: ScrollBehavior = 'smooth'): void => {
        const sidebar = codySidebarRef.current
        if (sidebar) {
            sidebar.scrollTo({
                behavior,
                top: sidebar.scrollHeight,
            })
        }
    }

    const onReset = useCallback(() => {
        reset()
        setShowHistory(false)
    }, [reset, setShowHistory])

    useEffect(() => {
        const sidebar = codySidebarRef.current
        if (sidebar && shouldScrollToBottom) {
            scrollToBottom('auto')
        }
    }, [transcript, shouldScrollToBottom, messageInProgress])

    return (
        <div className={styles.mainWrapper}>
            <div className={styles.codySidebar} ref={codySidebarRef} onScroll={handleScroll}>
                <div className={styles.codySidebarHeader}>
                    <div>
                        <Tooltip content="Start a new conversation">
                            <Button variant="icon" aria-label="Start a new conversation" onClick={onReset}>
                                <Icon aria-hidden={true} svgPath={mdiReload} />
                            </Button>
                        </Tooltip>
                    </div>
                    <div>
                        <CodyLogo />
                        Ask Cody
                    </div>
                    <div className="d-flex">
                        <Tooltip content="Chat history">
                            <Button
                                variant="icon"
                                className="mr-2"
                                aria-label="Chat history"
                                onClick={() => setShowHistory(showing => !showing)}
                            >
                                <Icon aria-hidden={true} svgPath={mdiHistory} />
                            </Button>
                        </Tooltip>
                        <Button variant="icon" aria-label="Close" onClick={onClose}>
                            <Icon aria-hidden={true} svgPath={mdiClose} />
                        </Button>
                    </div>
                </div>
                {showHistory ? (
                    <ChatHistory
                        transcriptHistory={transcriptHistory}
                        loadTranscript={loadTranscriptFromHistory}
                        closeHistory={() => setShowHistory(false)}
                        clearHistory={clearHistory}
                    />
                ) : (
                    <Chat
                        messageInProgress={messageInProgress}
                        messageBeingEdited={messageBeingEdited}
                        setMessageBeingEdited={setMessageBeingEdited}
                        transcript={transcript}
                        formInput={formInput}
                        setFormInput={setFormInput}
                        inputHistory={inputHistory}
                        setInputHistory={setInputHistory}
                        onSubmit={submitMessage}
                        contextStatus={getChatContext()}
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
                        editButtonOnSubmit={editMessage}
                        textAreaComponent={AutoResizableTextArea}
                        codeBlocksCopyButtonClassName={styles.codeBlocksCopyButton}
                        transcriptActionClassName={styles.transcriptAction}
                    />
                )}
            </div>
            {showScrollDownButton && <ScrollDownButton onClick={() => scrollToBottom('smooth')} />}
        </div>
    )
}

const ScrollDownButton = ({ onClick }: { onClick: () => void }): JSX.Element => (
    <div className={styles.scrollButtonWrapper}>
        <Button className={styles.scrollButton} onClick={onClick}>
            <Icon aria-label="Scroll down" svgPath={mdiArrowDown} />
        </Button>
    </div>
)

const EditButton: React.FunctionComponent<EditButtonProps> = ({
    className,
    messageBeingEdited,
    setMessageBeingEdited,
}) => (
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

const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = ({ className, disabled, onClick }) => (
    <button className={classNames(className, styles.submitButton)} type="submit" disabled={disabled} onClick={onClick}>
        <Icon aria-label="Submit" svgPath={mdiSend} />
    </button>
)

const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>

interface AutoResizableTextAreaProps extends ChatUITextAreaProps {}

export const AutoResizableTextArea: React.FC<AutoResizableTextAreaProps> = ({
    value,
    onInput,
    onKeyDown,
    className,
}) => {
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
        adjustTextAreaHeight()
    }, [adjustTextAreaHeight, value, width])

    return (
        <TextArea
            ref={textAreaRef}
            className={className}
            value={value}
            onChange={handleChange}
            rows={1}
            autoFocus={false}
            required={true}
            onKeyDown={onKeyDown}
            onInput={onInput}
        />
    )
}
