import type { FC } from 'react'

import { type Location, Navigate, type Params, useLocation, useParams } from 'react-router-dom'

interface RedirectRouteProps {
    getRedirectURL: (data: { location: Location; params: Params }) => string
}

export const RedirectRoute: FC<RedirectRouteProps> = props => {
    const { getRedirectURL } = props

    const location = useLocation()
    const params = useParams()

    return <Navigate to={getRedirectURL({ location, params })} replace={true} />
}
