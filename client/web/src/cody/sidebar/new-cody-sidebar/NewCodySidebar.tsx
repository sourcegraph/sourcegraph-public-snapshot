import { Suspense, type FC, useRef, useCallback, useState } from 'react'

import { mdiClose, mdiPlus, mdiArrowLeft, mdiHistory } from '@mdi/js'

import { CodyLogo } from '@sourcegraph/cody-ui'
import type { CodyWebChatContextClient } from '@sourcegraph/cody-web'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Alert, Button, H4, Icon, LoadingSpinner, ProductStatusBadge, Tooltip } from '@sourcegraph/wildcard'

import styles from './NewCodySidebar.module.scss'

const LazyCodySidebarWebChat = lazyComponent(() => import('./NewCodySidebarWebChat'), 'NewCodySidebarWebChat')

export interface Repository {
    id: string
    name: string
}

interface NewCodySidebarProps {
    filePath: string | undefined
    repository: Repository
    isAuthorized: boolean
    onClose: () => void
}

export const NewCodySidebar: FC<NewCodySidebarProps> = props => {
    const { repository, filePath, isAuthorized, onClose } = props

    const [chatMode, setChatMode] = useState<'chat' | 'history'>('chat')
    const codyClientRef = useRef<CodyWebChatContextClient>()

    const handleShowHistory = (): void => {
        setChatMode('history')
    }

    const handleShowChat = (): void => {
        setChatMode('chat')
    }

    const handleCreateNewChat = async (): Promise<void> => {
        if (codyClientRef.current) {
            await codyClientRef.current.createNewChat()
            setChatMode('chat')
        }
    }

    const handleSelectChat = (): void => {
        setChatMode('chat')
    }

    const handleClientCreation = useCallback((client: CodyWebChatContextClient): void => {
        codyClientRef.current = client
    }, [])

    return (
        <div className={styles.root}>
            <div className={styles.header}>
                <div className={styles.headerActions}>
                    {chatMode === 'history' && (
                        <Tooltip content="Go back to chat">
                            <Button variant="icon" aria-label="Create new chat" onClick={handleShowChat}>
                                <Icon aria-hidden={true} svgPath={mdiArrowLeft} />
                            </Button>
                        </Tooltip>
                    )}

                    {chatMode === 'chat' && (
                        <Tooltip content="Show chat history">
                            <Button variant="icon" aria-label="Show chat history" onClick={handleShowHistory}>
                                <Icon aria-hidden={true} svgPath={mdiHistory} />
                            </Button>
                        </Tooltip>
                    )}

                    <Tooltip content="Start a new chat">
                        <Button variant="icon" aria-label="Create new chat" onClick={handleCreateNewChat}>
                            <Icon aria-hidden={true} svgPath={mdiPlus} />
                        </Button>
                    </Tooltip>
                </div>

                <span className={styles.headerLogo}>
                    <CodyLogo />
                    Cody
                    <div className="ml-2">
                        <ProductStatusBadge status="beta" />
                    </div>
                </span>

                <Button variant="icon" aria-label="Close" onClick={onClose}>
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>

            {isAuthorized && (
                <Suspense
                    fallback={
                        <div className="flex flex-1 align-items-center m-2">
                            <LoadingSpinner className="mr-2" /> Loading Cody client
                        </div>
                    }
                >
                    <LazyCodySidebarWebChat
                        mode={chatMode}
                        filePath={filePath}
                        repository={repository}
                        onChatSelect={handleSelectChat}
                        onClientCreated={handleClientCreation}
                    />
                </Suspense>
            )}

            {!isAuthorized && (
                <Alert variant="info" className="m-3">
                    <H4>Cody is only available to signed-in users</H4>
                    Sign in to get access to use Cody
                </Alert>
            )}
        </div>
    )
}
