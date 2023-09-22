import * as React from 'react'

import { Routes, Route } from 'react-router-dom'

import { LoadingSpinner, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { RouteError } from '../components/ErrorBoundary'
import { NotFoundPage } from '../components/HeroPage'

import { SpongeLog } from './SpongeLog'

export interface Props {
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

const AuthenticatedSpongeLogsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    if (props.isSourcegraphDotCom) {
        return <NotFoundPage pageType="sponge log" />
    }
    return (
        <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
            <Routes>
                <Route path="" element={<Text>Spong logs landing page</Text>} errorElement={<RouteError />} />
                <Route path=":uuid" element={<SpongeLog />} errorElement={<RouteError />} />
                <Route path="*" element={<NotFoundPage pageType="sponge log" />} errorElement={<RouteError />} />
            </Routes>
        </React.Suspense>
    )
}

export const SpongeLogsArea = withAuthenticatedUser(AuthenticatedSpongeLogsArea)
