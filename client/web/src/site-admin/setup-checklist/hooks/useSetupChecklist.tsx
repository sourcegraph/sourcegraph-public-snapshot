import { gql, useQuery } from '@sourcegraph/http-client'
import { Code, Link, Text } from '@sourcegraph/wildcard'

import type { SetupChecklistResult, SetupChecklistVariables } from '../../../graphql-operations'

const QUERY = gql`
    query SetupChecklist {
        site {
            externalServicesCounts {
                remoteExternalServicesCount
            }
            productSubscription {
                license {
                    isValid
                    tags
                }
            }
        }
    }
`
export enum SetupChecklistItem {
    AddLicense = 'Add license',
    ConnectCodeHost = 'Connect code host',
}

interface UseSetupChecklistReturnType {
    data: { id: SetupChecklistItem; name: string; path: string; info: React.ReactNode; configured?: boolean }[]
    error?: any
    loading: boolean
}

export function useSetupChecklist(): UseSetupChecklistReturnType {
    const { data, loading, error } = useQuery<SetupChecklistResult, SetupChecklistVariables>(QUERY, {})

    const codeHostsCounts = data?.site?.externalServicesCounts?.remoteExternalServicesCount ?? 0
    const license = data?.site?.productSubscription?.license
    return {
        data: [
            {
                id: SetupChecklistItem.AddLicense,
                name: 'Add license',
                path: '/site-admin/configuration',
                info: (
                    <Text className="m-0">
                        Add a new <Code>licenseKey</Code> json field.{' '}
                        <Link to="/help/admin/config/site_config#licenseKey">Learn more</Link>
                    </Text>
                ),
                configured: license?.isValid && !license?.tags?.includes('plan:free-1'),
            },
            {
                id: SetupChecklistItem.ConnectCodeHost,
                name: 'Add remote repositories',
                path: '/site-admin/external-services/new',
                info: <div>Add a codehost and sync its repositories</div>,
                configured: codeHostsCounts > 0,
            },
        ],
        loading,
        error,
    }
}
