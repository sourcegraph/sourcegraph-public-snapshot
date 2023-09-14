import { useLocation } from 'react-router-dom'

import { gql, useQuery } from '@sourcegraph/http-client'
import { Link, LoadingSpinner, MenuLink } from '@sourcegraph/wildcard'

import type { AppUserConnectDotComAccountResult } from '../graphql-operations'

const QUERY = gql`
    query AppUserConnectDotComAccount {
        site {
            id
            appHasConnectedDotComAccount
        }
    }
`

export const AppUserConnectDotComAccount: React.FC = () => {
    const location = useLocation()

    const { data, loading } = useQuery<AppUserConnectDotComAccountResult, AppUserConnectDotComAccountResult>(QUERY, {
        nextFetchPolicy: 'cache-first',
    })

    const isAccountConnected = data?.site?.appHasConnectedDotComAccount

    const destination = encodeURIComponent(location.pathname + location.search)

    return loading ? (
        <MenuLink as={Link} to="#">
            <LoadingSpinner />
        </MenuLink>
    ) : !isAccountConnected ? (
        <MenuLink
            as={Link}
            to={`https://sourcegraph.com/user/settings/tokens/new/callback?requestFrom=APP&destination=${destination}`}
            target="_blank"
        >
            Connect to Sourcegraph.com
        </MenuLink>
    ) : null
}
