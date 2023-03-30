import { Routes, Route } from 'react-router-dom'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'

import { SentinelLandingPage } from './SentinelLandingPage'
import { SentinelView } from './SentinelView'

export interface SecurityAppRouterProps {
    authenticatedUser: AuthenticatedUser
}

export const SentinelAppRouter = withAuthenticatedUser<SecurityAppRouterProps>(() => (
    <Routes>
        <Route path="/demo" element={<SentinelView />} />
        <Route index={true} element={<SentinelLandingPage />} />
    </Routes>
))
