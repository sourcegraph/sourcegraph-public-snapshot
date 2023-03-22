import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { addSourcegraphAppOutboundUrlParameters } from '@sourcegraph/shared/src/util/url'
import { Link } from '@sourcegraph/wildcard'

import { LimitedAccessBanner } from '../../../../../components/LimitedAccessBanner'

interface Props {
    authenticatedUser: Pick<AuthenticatedUser, 'displayName' | 'emails'> | null | undefined
    className?: string
}
export const CodeInsightsLimitedAccessAppBanner: React.FC<Props> = props => (
    <LimitedAccessBanner storageKey="app.limitedAccessBannerDismissed.codeInsights" className={props.className}>
        Code Insights is currently available to try for free, up to 2 insights, while Sourcegraph App is in beta.
        Pricing and availability for Code Insights is subject to change in future releases.{' '}
        <strong>
            For unlimited access to Insights,{' '}
            <Link
                to={addSourcegraphAppOutboundUrlParameters(
                    'https://about.sourcegraph.com/get-started?app=enterprise',
                    'code-insights'
                )}
            >
                sign up for an Enterprise trial.
            </Link>
        </strong>
    </LimitedAccessBanner>
)
