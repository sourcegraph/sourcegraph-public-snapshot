import { differenceInDays, formatDistanceStrict } from 'date-fns'

import { gql, useQuery } from '@sourcegraph/http-client'
import { Code, Link, Text } from '@sourcegraph/wildcard'

import { SetupChecklistResult, SetupChecklistVariables } from '../../../graphql-operations'

const QUERY = gql`
    query SetupChecklist {
        site {
            externalServicesCounts {
                remoteExternalServicesCount
            }
            productSubscription {
                license {
                    isValid
                    expiresAt
                    tags
                    userCount
                }
            }
            users(deletedAt: { empty: true }) {
                totalCount
            }
        }
    }
`
export enum SetupChecklistItem {
    AddLicense = 'Add license',
    ConnectCodeHost = 'Connect code host',
}

export interface SetupChecklistItemType {
    id: SetupChecklistItem
    name: string
    path: string
    info: React.ReactNode
    notification?: {
        text: string
        type: 'warning' | 'danger'
    }
}
interface UseSetupChecklistReturnType {
    data: SetupChecklistItemType[]
    error?: any
    loading: boolean
}

const NOTIFY_LICENSE_EXPIRATION_DAYS = 7

/**
 * Returns either text of why setup/action is required or undefined if everything is good
 *
 * @param args
 * @param now for testing purposes
 */
export const getLicenseSetupStatus = (
    args?: Pick<SetupChecklistResult['site'], 'users' | 'productSubscription'>,
    now = new Date()
): SetupChecklistItemType['notification'] => {
    const license = args?.productSubscription?.license

    if (!license?.isValid) {
        return { text: 'The Sourcegraph license key is invalid.', type: 'danger' }
    }

    if (license?.tags?.includes('dev')) {
        return
    }

    const expiresAt = license?.expiresAt ? new Date(license.expiresAt) : undefined
    if (expiresAt && differenceInDays(expiresAt, now) <= NOTIFY_LICENSE_EXPIRATION_DAYS) {
        const isExpired = differenceInDays(expiresAt, now) <= 0
        return {
            text: `The Sourcegraph license ${
                isExpired
                    ? 'expired ' + formatDistanceStrict(expiresAt, now) + ' ago. Please, get a new license.' // 'Expired two months ago'
                    : 'will expire in ' + formatDistanceStrict(expiresAt, now) + '. Please, renew it soon.' // 'Will expire in two months'
            }`,
            type: isExpired ? 'danger' : 'warning',
        }
    }

    if (license?.tags?.includes('plan:free-1')) {
        return {
            text: 'You are on a free plan. Please, upgrade your license to unlock more features.',
            type: 'warning',
        }
    }

    const userCount = args?.users?.totalCount
    const hasExceededUserCount =
        !!userCount && !!license?.userCount && userCount > license?.userCount && !license?.tags.includes('true-up')
    if (hasExceededUserCount) {
        return { text: 'Your user count has exceeded your license limit. Please, get a new license.', type: 'danger' }
    }

    return
}

const getCodehostSetupStatus = (
    args?: Pick<SetupChecklistResult['site'], 'externalServicesCounts'>
): SetupChecklistItemType['notification'] => {
    const codeHostsCounts = args?.externalServicesCounts?.remoteExternalServicesCount
    // todo: add more checks such as valid code host connections
    if (codeHostsCounts === 0) {
        return { text: 'Add code host connection', type: 'danger' }
    }
    return
}

export function useSetupChecklist(): UseSetupChecklistReturnType {
    const { data, loading, error } = useQuery<SetupChecklistResult, SetupChecklistVariables>(QUERY, {})

    return {
        data: [
            {
                id: SetupChecklistItem.AddLicense,
                name: 'Add license',
                path: '/site-admin/configuration',
                notification: getLicenseSetupStatus(data?.site),
                info: (
                    <Text className="m-0">
                        Add a new <Code>licenseKey</Code> json field.{' '}
                        <Link to="/help/admin/config/site_config#licenseKey">Learn more</Link>
                    </Text>
                ),
            },
            {
                id: SetupChecklistItem.ConnectCodeHost,
                name: 'Connect codehost(s)',
                path: '/site-admin/external-services/new',
                info: <div>Add codehost connections and sync its repositories</div>,
                notification: getCodehostSetupStatus(data?.site),
            },
        ],
        loading,
        error,
    }
}
