import type { FC } from 'react'

import { Navigate } from 'react-router-dom'

import { PageRoutes } from './routes.constants'

export const IndexPage: FC = () => (
    <Navigate
        replace={true}
        to={window.context?.codeSearchEnabledOnInstance ? PageRoutes.Search : PageRoutes.CodyDashboard}
    />
)
