import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'

export interface CampaignsSiteAdminMarketingPageProps {}

export const CampaignsSiteAdminMarketingPage: React.FunctionComponent<CampaignsSiteAdminMarketingPageProps> = () => (
    <CampaignsMarketing
        body={
            <section className="my-3">
                <h2>Get started</h2>
                <ol>
                    <li>
                        See <a href="https://docs.sourcegraph.com/user/campaigns/getting_started">Getting started</a> to
                        enable campaigns on your Sourcegraph instance.
                    </li>
                    <li>
                        Now,{' '}
                        <a href="https://docs.sourcegraph.com/user/campaigns/creating_campaign_from_patches">
                            create your first campaign
                        </a>
                        .
                    </li>
                </ol>
                <div className="alert alert-info mt-3">
                    <p>
                        <strong>
                            By enabling this feature, you agree to test out functionality that is currently in beta, and
                            acknowledge the following:
                        </strong>
                    </p>
                    <ul className="mb-3">
                        <li>
                            During the beta period, campaigns are free to use. After the beta period ends, campaigns
                            will be available as a paid add-on.
                        </li>
                        <li>
                            Campaigns are not included as part of the Enterprise tiers (Enterprise, Enterprise Starter,
                            Enterprise Plus, etc.) license tiers.
                        </li>
                    </ul>
                </div>
            </section>
        }
    />
)
