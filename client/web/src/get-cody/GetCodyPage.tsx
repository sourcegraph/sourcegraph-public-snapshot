import { useState } from 'react'

import { useNavigate, useLocation } from 'react-router-dom'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

interface GetCodyPageProps {
    authenticatedUser: AuthenticatedUser | null
}

export const GetCodyPage: React.FunctionComponent<GetCodyPageProps> = ({ authenticatedUser }) => {
    const navigate = useNavigate()
    const location = useLocation()
    const [search] = useState(location.search)

    if (authenticatedUser) {
        navigate(`/cody/manage${search || ''}`)
    } else {
        window.location.href = '/cody'
    }

    return <></>
}
