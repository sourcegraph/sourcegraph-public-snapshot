import { useEffect, useRef, useState } from 'react'

import { mdiClose, mdiSend } from '@mdi/js'

import { Button, Icon } from '@sourcegraph/wildcard'

import { useChatStoreState } from '../../web/src/stores/codyChat'

import { Chat, ChatUISubmitButtonProps } from './Chat'
import { FileLinkProps } from './chat/ContextFiles'
import { CodyLogo } from './icons/CodyLogo'
import { Terms } from './Terms'

import styles from './CodyChat.module.scss'

export const SCROLL_THRESHOLD = 40

interface CodyChatProps {
    repoName: string
    onClose: () => void
}

export const CodyChat = ({ repoName, onClose }: CodyChatProps): JSX.Element => {
    const { onSubmit, messageInProgress, transcript } = useChatStoreState()

    const codySidebarRef = useRef<HTMLDivElement>(null)
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])
    const [shouldScrollToBottom, setShouldScrollToBottom] = useState(true)

    const chatTitle = 'Ask Cody'

    const handleScroll = () => {
        const sidebar = codySidebarRef.current
        if (sidebar) {
            const scrollOffset = sidebar.scrollHeight - sidebar.scrollTop - sidebar.clientHeight
            setShouldScrollToBottom(scrollOffset <= SCROLL_THRESHOLD)
        }
    }

    useEffect(() => {
        // Only scroll if the user didn't scroll up manually more than the scrolling threshold.
        // That is so that you can freely copy content or read up on older content while new
        // content is being produced.
        const sidebar = codySidebarRef.current
        if (sidebar && shouldScrollToBottom) {
            sidebar.scrollTo({ behavior: 'smooth', top: sidebar.scrollHeight })
        }
    }, [transcript, shouldScrollToBottom])

    return (
        <div className={styles.codySidebar} ref={codySidebarRef} onScroll={handleScroll}>
            <div className={styles.codySidebarHeader}>
                <div>
                    <CodyLogo />
                    {chatTitle}
                </div>
                <div>
                    <Button variant="icon" aria-label="Close" onClick={onClose}>
                        <Icon aria-hidden={true} svgPath={mdiClose} />
                    </Button>
                </div>
            </div>
            <Chat
                messageInProgress={messageInProgress}
                transcript={transcript}
                formInput={formInput}
                setFormInput={setFormInput}
                inputHistory={inputHistory}
                setInputHistory={setInputHistory}
                onSubmit={onSubmit}
                submitButtonComponent={SubmitButton}
                fileLinkComponent={FileLink}
                className={styles.container}
                afterTips={
                    <details className="small mt-2">
                        <summary>Terms</summary>
                        <Terms />
                    </details>
                }
                bubbleContentClassName={styles.bubbleContent}
                humanBubbleContentClassName={styles.humanBubbleContent}
                botBubbleContentClassName={styles.botBubbleContent}
                bubbleFooterClassName="text-muted small"
                bubbleLoaderDotClassName={styles.bubbleLoaderDot}
                inputRowClassName={styles.inputRow}
                chatInputClassName={styles.chatInput}
            />
        </div>
    )
}

const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = ({ className, disabled, onClick }) => (
    <button className={className} type="submit" disabled={disabled} onClick={onClick}>
        <Icon svgPath={mdiSend} />
    </button>
)

const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>
