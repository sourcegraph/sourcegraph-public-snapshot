import { FC } from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'

const SecurityAppLazyRouter = lazyComponent(() => import('./SecurityAppRouter'), 'SecurityAppRouter')

export interface SecurityRouterProps {
    authenticatedUser: AuthenticatedUser | null
}

export const SecurityRouter: FC<SecurityRouterProps> = props => {
    const { authenticatedUser } = props

    return <SecurityAppLazyRouter authenticatedUser={authenticatedUser} />
}
