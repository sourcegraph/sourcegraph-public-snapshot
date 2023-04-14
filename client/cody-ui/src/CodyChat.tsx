import { useCallback, useEffect, useRef, useState } from 'react'

import { mdiClose, mdiSend, mdiArrowDown } from '@mdi/js'

import { Button, Icon } from '@sourcegraph/wildcard'

import { useChatStoreState } from '../../web/src/stores/codyChat'

import { Chat, ChatUISubmitButtonProps } from './Chat'
import { FileLinkProps } from './chat/ContextFiles'
import { CodyLogo } from './icons/CodyLogo'
import { Terms } from './Terms'

import styles from './CodyChat.module.scss'

export const SCROLL_THRESHOLD = 100

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
    const [showScrollDownButton, setShowScrollDownButton] = useState(false)

    const chatTitle = 'Ask Cody'

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

    return (
        <div className={styles.mainWrapper}>
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
            {showScrollDownButton && <ScrollDownButton onClick={() => scrollToBottom('smooth')} />}
        </div>
    )
}

const ScrollDownButton = ({ onClick }: { onClick: () => void }): JSX.Element => (
    <div className={styles.scrollButtonWrapper}>
        <div className={styles.scrollButton} onClick={onClick}>
            <Icon svgPath={mdiArrowDown} />
        </div>
    </div>
)

const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = ({ className, disabled, onClick }) => (
    <button className={className} type="submit" disabled={disabled} onClick={onClick}>
        <Icon svgPath={mdiSend} />
    </button>
)

const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>
