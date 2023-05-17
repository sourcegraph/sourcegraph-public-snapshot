import { useCallback, useEffect, useRef, useState } from 'react'

import { mdiClose, mdiFormatListBulleted, mdiPlus, mdiDelete } from '@mdi/js'

import { CodyLogo } from '@sourcegraph/cody-ui/src/icons/CodyLogo'
import { Button, Icon, Tooltip, Badge } from '@sourcegraph/wildcard'

import { ChatUI, ScrollDownButton } from '../components/ChatUI'
import { HistoryList } from '../components/HistoryList'
import { useChatStoreState } from '../stores/chat'

import styles from './CodySidebar.module.scss'

export const SCROLL_THRESHOLD = 100

interface CodySidebarProps {
    onClose?: () => void
}

export const CodySidebar = ({ onClose }: CodySidebarProps): JSX.Element => {
    const { reset, transcript, messageInProgress, clearHistory } = useChatStoreState()

    const codySidebarRef = useRef<HTMLDivElement>(null)
    const [showHistory, setShowHistory] = useState(false)
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

    const onReset = useCallback(() => {
        reset()
        setShowHistory(false)
    }, [reset, setShowHistory])

    const closeHistory = useCallback(() => {
        setShowHistory(false)
    }, [setShowHistory])

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
                    <div className="d-flex col-3 p-0">
                        <Tooltip content="Chat history">
                            <Button
                                variant="icon"
                                className="mr-2"
                                aria-label="Active chats"
                                onClick={() => setShowHistory(showing => !showing)}
                            >
                                <Icon aria-hidden={true} svgPath={mdiFormatListBulleted} />
                            </Button>
                        </Tooltip>
                        <Tooltip content="Start a new chat">
                            <Button variant="icon" aria-label="Start a new chat" onClick={onReset}>
                                <Icon aria-hidden={true} svgPath={mdiPlus} />
                            </Button>
                        </Tooltip>
                        {showHistory && (
                            <Tooltip content="Clear all chats">
                                <Button
                                    variant="icon"
                                    className="ml-2"
                                    aria-label="Clear all chats"
                                    onClick={clearHistory}
                                >
                                    <Icon aria-hidden={true} svgPath={mdiDelete} />
                                </Button>
                            </Tooltip>
                        )}
                    </div>
                    <div className="col-6 d-flex justify-content-center">
                        <CodyLogo />
                        {showHistory ? 'Chats' : 'Ask Cody'}
                        <div className="ml-2">
                            <Badge variant="info">Beta</Badge>
                        </div>
                    </div>
                    <div className="col-3 d-flex justify-content-end p-0">
                        <Button variant="icon" aria-label="Close" onClick={onClose}>
                            <Icon aria-hidden={true} svgPath={mdiClose} />
                        </Button>
                    </div>
                </div>

                {showHistory ? <HistoryList onSelect={closeHistory} itemClassName="rounded-0" /> : <ChatUI />}
            </div>
            {showScrollDownButton && <ScrollDownButton onClick={() => scrollToBottom('smooth')} />}
        </div>
    )
}
