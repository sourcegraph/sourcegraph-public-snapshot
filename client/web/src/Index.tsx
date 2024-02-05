import type { FC } from 'react'

import { Navigate } from 'react-router-dom'

import { PageRoutes } from './routes.constants'
import { isCodyOnlyLicense } from './util/license'

export const IndexPage: FC = () => {
    let redirectRoute = PageRoutes.Search
    if (isCodyOnlyLicense()) {
        redirectRoute = PageRoutes.Cody
    }

    return <Navigate replace={true} to={redirectRoute} />
}
