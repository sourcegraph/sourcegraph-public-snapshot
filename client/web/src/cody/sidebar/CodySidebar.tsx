import { useCallback, useEffect, useRef, useState } from 'react'

import { mdiClose, mdiReload, mdiHistory } from '@mdi/js'

import { CodyLogo } from '@sourcegraph/cody-ui/src/icons/CodyLogo'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { ChatUI, ScrollDownButton } from '../components/ChatUI'
import { useChatStoreState } from '../stores/chat'

import { History } from './History'

import styles from './CodySidebar.module.scss'

export const SCROLL_THRESHOLD = 100

interface CodySidebarProps {
    onClose?: () => void
}

export const CodySidebar = ({ onClose }: CodySidebarProps): JSX.Element => {
    const { reset, transcript, messageInProgress } = useChatStoreState()

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

    useEffect(() => {
        const sidebar = codySidebarRef.current
        if (sidebar && shouldScrollToBottom) {
            scrollToBottom('auto')
        }
    }, [transcript, shouldScrollToBottom, messageInProgress])

    const closeHistory = useCallback(() => setShowHistory(false), [setShowHistory])

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

                {showHistory ? <History closeHistory={closeHistory} /> : <ChatUI />}
            </div>
            {showScrollDownButton && <ScrollDownButton onClick={() => scrollToBottom('smooth')} />}
        </div>
    )
}
