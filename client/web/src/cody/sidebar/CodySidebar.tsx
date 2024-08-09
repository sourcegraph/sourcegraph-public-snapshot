import { Suspense, type FC } from 'react'

import { mdiClose } from '@mdi/js'

import { CodyLogo } from '@sourcegraph/cody-ui'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Alert, Button, H4, Icon, LoadingSpinner, ProductStatusBadge } from '@sourcegraph/wildcard'

import styles from './CodySidebar.module.scss'

const LazyCodySidebarWebChat = lazyComponent(() => import('./CodySidebarWebChat'), 'CodySidebarWebChat')

export interface Repository {
    id: string
    name: string
}

interface CodySidebarProps {
    filePath: string | undefined
    repository: Repository
    authenticatedUser: AuthenticatedUser | null
    onClose: () => void
}

export const CodySidebar: FC<CodySidebarProps> = props => {
    const { repository, filePath, authenticatedUser, onClose } = props

    return (
        <div className={styles.root}>
            <div className={styles.header}>
                <div />
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

            {authenticatedUser && (
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

            {!authenticatedUser && (
                <Alert variant="info" className="m-3">
                    <H4>Cody is only available to signed-in users</H4>
                    Sign in to get access to use Cody
                </Alert>
            )}
        </div>
    )
}
