import { useQuery } from '@sourcegraph/http-client'

import { ExternalAccountFields, MinExternalAccountsVariables } from '../graphql-operations'
import { USER_EXTERNAL_ACCOUNTS } from '../user/settings/backend'

type MinExternalAccount = Pick<ExternalAccountFields, 'id' | 'serviceID' | 'serviceType' | 'accountData'>

interface UserExternalAccountsResult {
    user: {
        externalAccounts: {
            nodes: MinExternalAccount[]
        }
    }
}

export function useUserExternalAccounts(username: string): { data: MinExternalAccount[]; loading: boolean } {
    const { data, loading } = useQuery<UserExternalAccountsResult, MinExternalAccountsVariables>(
        USER_EXTERNAL_ACCOUNTS,
        {
            variables: { username },
        }
    )

    if (!data) {
        return { data: [], loading }
    }

    return { data: data?.user?.externalAccounts.nodes, loading }
}
