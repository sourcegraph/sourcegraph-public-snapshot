import * as React from 'react'
import { SingleValueCard } from '../components/SingleValueCard'
import { numberWithCommas } from '../util/strings'
import { formatUserCount } from './helpers'

interface Props {
    actualUserCount: number
    userCount: number
    actualUserCountDate: string | null
}
/**
 * Displays a summary of the site's true-up pricing status.
 */
export const TrueUpStatusSummary: React.FunctionComponent<Props> = ({
    actualUserCount,
    actualUserCountDate,
    userCount,
}) => (
    <>
        <div className="true-up-status-summary mb-2 mt-4">
            <div className="true-up-status-summary__container">
                <SingleValueCard
                    className="true-up-status-summary__item"
                    value={numberWithCommas(userCount)}
                    valueTooltip={`${formatUserCount(userCount, true)} license`}
                    title="Licensed users"
                    subText="The number of users that are currently covered by your license. The true-up model allows having more users, and additional users will incur a retroactive charge on renewal."
                />
                <SingleValueCard
                    className="true-up-status-summary__item"
                    value={numberWithCommas(actualUserCount)}
                    valueTooltip={`${numberWithCommas(actualUserCount)} total users${actualUserCountDate !== '' &&
                        `(reached on ${actualUserCountDate})`}`}
                    title="Maximum users"
                    subText="This is the highest peak of users on your installation since the license started, and this is the minimum number you need to purchase when you renew your license."
                />
                <SingleValueCard
                    className="true-up-status-summary__item"
                    value={numberWithCommas(Math.max(0, actualUserCount - userCount))}
                    valueTooltip={`${numberWithCommas(
                        Math.max(0, actualUserCount - userCount)
                    )} users over${actualUserCountDate !== '' && `(on ${actualUserCountDate})`}`}
                    title="Users over license"
                    subText="The true-up model has a retroactive charge for these users at the next renewal. If you want to update your license sooner to prevent this, please contact sales@sourcegraph.com."
                    valueClassName={userCount - actualUserCount < 0 ? 'text-danger' : ''}
                />
            </div>
            <small>
                Learn more about{' '}
                <a href="https://about.sourcegraph.com/pricing" target="_blank">
                    Sourcegraph's true-up pricing model
                </a>
                .
            </small>
        </div>
    </>
)
