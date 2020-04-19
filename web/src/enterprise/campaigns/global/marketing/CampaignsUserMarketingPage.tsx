import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'

export interface CampaignsUserMarketingPageProps {
    /** When true, the message shown will indicate that the user can just not use it because read-access has not been provided to non site-admins. */
    enableReadAccess: boolean
}

export const CampaignsUserMarketingPage: React.FunctionComponent<CampaignsUserMarketingPageProps> = ({
    enableReadAccess,
}) => (
    <CampaignsMarketing
        body={
            <section className="my-3">
                <h2>Interested in campaigns?</h2>
                <p>
                    At this time, creating and managing campaigns is only available to Sourcegraph admins. What can you
                    do?
                </p>
                <div className="row">
                    <ol>
                        <li>Let your Sourcegraph admin know you're interested in using campaigns for your team.</li>
                        {enableReadAccess && (
                            <li>
                                Ask your Sourcegraph admin for{' '}
                                <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener">
                                    read-only access
                                </a>{' '}
                                to campaigns.
                                <div className="alert alert-info mt-3">
                                    <b>NOTE:</b> Repository permissions are NOT enforced by campaigns, so your admin may
                                    not grant read-only access if your Sourcegraph instance has repository permissions
                                    configured.
                                </div>
                            </li>
                        )}
                        <li>
                            Learn how to{' '}
                            <a href="https://docs.sourcegraph.com/user/campaigns#creating-campaigns">
                                get started creating campaigns
                            </a>
                            .
                        </li>
                    </ol>
                </div>

                <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener" className="btn btn-primary">
                    Read more about campaigns
                </a>
            </section>
        }
    />
)
