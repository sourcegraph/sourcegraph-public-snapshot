import type { FC } from 'react'

import classNames from 'classnames'
import { formatRelative, parseISO } from 'date-fns'

import { useQuery, gql } from '@sourcegraph/http-client'
import { Text, Container, PageHeader, LoadingSpinner, ErrorAlert } from '@sourcegraph/wildcard'

import type { CodyGatewayRateLimitStatusResult } from '../../../../graphql-operations'

import styles from './RateLimitsTab.module.scss'

const formatDate = (date: string | null): string => {
    if (!date) {
        return ''
    }
    return ` ${formatRelative(parseISO(date), new Date())}`
}

interface RateLimitStatus {
    feature: string
    limit: string
    usage: string
    nextLimitReset: string | null
}

export const GET_CODY_RATE_LIMIT_STATUS = gql`
    query CodyGatewayRateLimitStatus {
        site {
            codyGatewayRateLimitStatus {
                feature
                limit
                usage
                nextLimitReset
            }
        }
    }
`
interface RateLimitsTabProps {
    className?: string
}
export const RateLimitsTab: FC<RateLimitsTabProps> = props => {
    const { className } = props

    const { data, loading, error } = useQuery<CodyGatewayRateLimitStatusResult>(GET_CODY_RATE_LIMIT_STATUS, {
        fetchPolicy: 'network-only',
    })

    return (
        <div className="w-100">
            <PageHeader
                headingElement="h2"
                description="View current usage"
                path={[{ text: 'Usage Limits' }]}
                className="mb-3"
            />
            <div className={classNames(className, styles.root)}>
                <Container className={styles.container}>
                    {error && <ErrorAlert error={error} />}

                    {!error && loading && <LoadingSpinner />}
                    {!error && data && data.site.codyGatewayRateLimitStatus?.length === 0 && <NoRateLimitState />}
                    {!error &&
                        data?.site.codyGatewayRateLimitStatus &&
                        data.site.codyGatewayRateLimitStatus.length > 0 && (
                            <RateLimitTable limits={data.site.codyGatewayRateLimitStatus} />
                        )}
                </Container>
            </div>
        </div>
    )
}

interface RateLimitTableProps {
    limits: RateLimitStatus[]
}
const RateLimitTable: FC<RateLimitTableProps> = ({ limits }) => (
    <table className="table">
        <thead>
            <tr>
                <th>Feature</th>
                <th>Used</th>
                <th>Allowed</th>
                <th>Next reset</th>
            </tr>
        </thead>
        <tbody>
            {limits.map(limit => (
                <RateLimitRow key={limit.feature} limit={limit} />
            ))}
        </tbody>
    </table>
)

interface RateLimitRowProps {
    limit: RateLimitStatus
}
const RateLimitRow: FC<RateLimitRowProps> = ({ limit }) => (
    <tr>
        <td>
            <Text>{limit.feature}</Text>
        </td>
        <td>
            <Text>{limit.usage}</Text>
        </td>
        <td>
            <Text>{limit.limit}</Text>
        </td>
        <td>{limit.nextLimitReset && <Text>{formatDate(limit.nextLimitReset)}</Text>}</td>
    </tr>
)

export const NoRateLimitState: FC = () => (
    <div>
        <Text className="mb-0">No usage limits</Text>
    </div>
)
