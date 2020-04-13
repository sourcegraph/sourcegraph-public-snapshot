import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'
import { TelemetryProps } from '../../../../../../shared/src/telemetry/telemetryService'

export interface CampaignsUserMarketingPageProps extends TelemetryProps {
    /** When true, the message shown will indicate that the user can just not use it because read-access has not been provided to non site-admins. */
    enableReadAccess: boolean
}

export const CampaignsUserMarketingPage: React.FunctionComponent<CampaignsUserMarketingPageProps> = ({
    enableReadAccess,
    telemetryService,
}) => (
    <>
        <CampaignsMarketing
            telemetryService={telemetryService}
            subTitle={
                <i>
                    Ask your admin{' '}
                    {enableReadAccess && <a href="#getting-started">to enable read-access for campaigns</a>}
                    {!enableReadAccess && <a href="#getting-started">to enable campaigns for your team</a>}
                </i>
            }
        />

        <section className="my-3 text-center" id="getting-started">
            {!enableReadAccess && <h1>Ask your admin to enable campaigns for your team</h1>}
            {enableReadAccess && <h1>Ask your admin to enable read-access for campaigns</h1>}

            <p className="lead">
                Running Campaigns is currently only possible for site admins of your Sourcegraph instance.{' '}
                {enableReadAccess && (
                    <>
                        However, your administrator can{' '}
                        <a href="https://docs.sourcegraph.com/user/campaigns#configuration" rel="noopener">
                            enable read-only access to campaigns
                        </a>{' '}
                        for other users, by setting <code>campaigns.readAccess.enabled</code>.
                    </>
                )}
                {!enableReadAccess && (
                    <>
                        Your admin will need to{' '}
                        <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener">
                            update the configuration settings
                        </a>{' '}
                        to make large-scale code changes across many repositories and different code hosts.
                    </>
                )}
            </p>
            <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener" className="btn btn-primary">
                Learn how to get started with campaigns
            </a>
        </section>
    </>
)
