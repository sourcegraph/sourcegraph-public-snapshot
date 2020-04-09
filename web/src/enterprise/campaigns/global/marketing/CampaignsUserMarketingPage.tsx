import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'

export interface CampaignsUserMarketingPageProps {
    /** When true, the message shown will indicate that the user can just not use it because read-access has not been provided to non site-admins. */
    enableReadAccess: boolean
}

export const CampaignsUserMarketingPage: React.FunctionComponent<CampaignsUserMarketingPageProps> = ({
    enableReadAccess,
}) => (
    <>
        <CampaignsMarketing />

        <section className="my-3 text-center">
            <h1>Start using campaigns today.</h1>

            <p className="lead">
                Running Campaigns is currently only possible for site admins of your Sourcegraph instance.
                {enableReadAccess && (
                    <>
                        However, your administrator can enable read-only access to campaigns for other users, by setting{' '}
                        <code>campaigns.readAccess.enabled</code>.
                    </>
                )}
                {!enableReadAccess && <>Reach out to your Sourcegraph instance administrator so they can try it out.</>}
            </p>
        </section>
    </>
)
