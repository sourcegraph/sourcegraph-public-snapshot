import { Suspense, type FC } from 'react'

import { mdiClose } from '@mdi/js'

import { CodyLogo } from '@sourcegraph/cody-ui'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Button, Icon, Badge, LoadingSpinner, H4, Alert } from '@sourcegraph/wildcard'

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

    return (
        <div className={styles.root}>
            <div className={styles.header}>
                <div className="d-flex flex-shrink-0 align-items-center">
                    <CodyLogo />
                    Cody Web
                    <div className="ml-2">
                        <Badge variant="info">Experimental</Badge>
                    </div>
                </div>
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
                    <LazyCodySidebarWebChat filePath={filePath} repository={repository} />
                </Suspense>
            )}

            {!isAuthorized && (
                <Alert variant="info" className="m-3">
                    <H4>Cody Web is only available to signed-in users</H4>
                    Sign in to get access to use Cody Web
                </Alert>
            )}
        </div>
    )
}
