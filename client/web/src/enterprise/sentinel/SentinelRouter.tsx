import { FC } from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'

const SentinelAppLazyRouter = lazyComponent(() => import('./SentinelAppRouter'), 'SentinelAppRouter')

export interface SentinelRouterProps {
    authenticatedUser: AuthenticatedUser | null
}

export const SentinelRouter: FC<SentinelRouterProps> = ({ authenticatedUser }) => (
    <SentinelAppLazyRouter authenticatedUser={authenticatedUser} />
)
