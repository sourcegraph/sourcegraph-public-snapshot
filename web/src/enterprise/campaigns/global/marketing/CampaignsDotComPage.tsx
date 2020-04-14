import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'

export interface CampaignsDotComPageProps {}

export const CampaignsDotComPage: React.FunctionComponent<CampaignsDotComPageProps> = () => (
    <CampaignsMarketing
        body={
            <section className="my-3">
                <h2>Get started</h2>
                <p>
                    Campaigns are not available on Sourcegraph.com. Instead, use a private Sourcegraph instance to try
                    Campaigns.
                </p>
                <ol>
                    <li>
                        Install a private Sourcegraph instance using the{' '}
                        <a href="https://docs.sourcegraph.com/#quickstart-guide" rel="noopener">
                            quickstart guide.
                        </a>
                    </li>
                    <li>
                        <a href="https://docs.sourcegraph.com/admin/repo/add">Add repositories</a> from your code host
                        to Sourcegraph.
                    </li>
                    <li>
                        <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener">
                            Update the site configuration settings
                        </a>{' '}
                        to enable Campaigns.
                    </li>
                </ol>

                <a href="https://docs.sourcegraph.com/admin/install" rel="noopener" className="btn btn-primary">
                    Get started now
                </a>
                <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener" className="btn btn-primary ml-2">
                    Read more about Campaigns
                </a>
            </section>
        }
    />
)
