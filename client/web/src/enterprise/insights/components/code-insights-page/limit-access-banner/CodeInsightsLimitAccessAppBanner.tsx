import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { addSourcegraphAppOutboundUrlParameters, buildCloudTrialURL } from '@sourcegraph/shared/src/util/url'
import { Link } from '@sourcegraph/wildcard'

import { LimitedAccessBanner } from '../../../../../components/LimitedAccessBanner'

interface Props {
    authenticatedUser: Pick<AuthenticatedUser, 'displayName' | 'emails'> | null | undefined
}
export const CodeInsightsLimitedAccessAppBanner: React.FC<Props> = props => (
    <LimitedAccessBanner storageKey="app.limitedAccessBannerDismissed.codeInsights">
        Code Insights is currently available to try for free, up to 2 insights, while Sourcegraph App is in beta.
        Pricing and availability for Code Insights is subject to change in future releases.{' '}
        <strong>
            For unlimited access to Insights,{' '}
            <Link
                to={addSourcegraphAppOutboundUrlParameters(
                    buildCloudTrialURL(props.authenticatedUser),
                    'code-insights'
                )}
            >
                sign up for a Cloud Trial.
            </Link>
        </strong>
    </LimitedAccessBanner>
)
