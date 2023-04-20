import { useCallback, useEffect, useRef, useState } from 'react'

import { mdiClose, mdiSend, mdiArrowDown, mdiReload } from '@mdi/js'
import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { Chat, ChatUISubmitButtonProps, ChatUITextAreaProps } from '@sourcegraph/cody-ui/src/Chat'
import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'
import { CodyLogo } from '@sourcegraph/cody-ui/src/icons/CodyLogo'
import { CODY_TERMS_MARKDOWN } from '@sourcegraph/cody-ui/src/terms'
import { useQuery } from '@sourcegraph/http-client'
import { Button, ErrorAlert, Icon, LoadingSpinner, Text, TextArea, Tooltip } from '@sourcegraph/wildcard'

import { RepoEmbeddingExistsQueryResult, RepoEmbeddingExistsQueryVariables } from '../graphql-operations'
import { REPO_EMBEDDING_EXISTS_QUERY } from '../repo/repoRevisionSidebar/cody/backend'
import { useChatStoreState } from '../stores/codyChat'

import styles from './CodyChat.module.scss'

export const SCROLL_THRESHOLD = 100

interface CodyChatProps {
    onClose: () => void
}

export const CodyChat = ({ onClose }: CodyChatProps): JSX.Element => {
    const { reset, submitMessage, messageInProgress, transcript, repo } = useChatStoreState()

    const codySidebarRef = useRef<HTMLDivElement>(null)
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])
    const [shouldScrollToBottom, setShouldScrollToBottom] = useState(true)
    const [showScrollDownButton, setShowScrollDownButton] = useState(false)

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

    useEffect(() => {
        const sidebar = codySidebarRef.current
        if (sidebar && shouldScrollToBottom) {
            scrollToBottom('auto')
        }
    }, [transcript, shouldScrollToBottom, messageInProgress])

    const {
        data: embeddingExistsData,
        loading: embeddingExistsLoading,
        error: embeddingExistsError,
    } = useQuery<RepoEmbeddingExistsQueryResult, RepoEmbeddingExistsQueryVariables>(REPO_EMBEDDING_EXISTS_QUERY, {
        variables: { repoName: repo },
    })

    const shouldShowResetButton =
        !embeddingExistsLoading && !embeddingExistsError && embeddingExistsData?.repository?.embeddingExists

    return (
        <div className={styles.mainWrapper}>
            <div className={styles.codySidebar} ref={codySidebarRef} onScroll={handleScroll}>
                <div className={styles.codySidebarHeader}>
                    {shouldShowResetButton && (
                        <div>
                            <Tooltip content="Start a new conversation">
                                <Button variant="icon" aria-label="Start a new conversation" onClick={reset}>
                                    <Icon aria-hidden={true} svgPath={mdiReload} />
                                </Button>
                            </Tooltip>
                        </div>
                    )}
                    <div>
                        <CodyLogo />
                        Ask Cody
                    </div>
                    <div>
                        <Button variant="icon" aria-label="Close" onClick={onClose}>
                            <Icon aria-hidden={true} svgPath={mdiClose} />
                        </Button>
                    </div>
                </div>
                {embeddingExistsLoading ? (
                    <LoadingSpinner className="m-3" />
                ) : embeddingExistsError ? (
                    <ErrorAlert error={embeddingExistsError} className="m-3" />
                ) : !embeddingExistsData?.repository?.embeddingExists ? (
                    <Text className="m-3">Repository embeddings are not available.</Text>
                ) : (
                    <Chat
                        messageInProgress={messageInProgress}
                        transcript={transcript}
                        formInput={formInput}
                        setFormInput={setFormInput}
                        inputHistory={inputHistory}
                        setInputHistory={setInputHistory}
                        onSubmit={submitMessage}
                        submitButtonComponent={SubmitButton}
                        fileLinkComponent={FileLink}
                        className={styles.container}
                        afterTips={CODY_TERMS_MARKDOWN}
                        transcriptItemClassName={styles.transcriptItem}
                        humanTranscriptItemClassName={styles.humanTranscriptItem}
                        transcriptItemParticipantClassName="text-muted"
                        inputRowClassName={styles.inputRow}
                        chatInputClassName={styles.chatInput}
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
