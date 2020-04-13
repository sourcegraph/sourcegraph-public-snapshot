import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'

export interface CampaignsDotComPageProps {}

export const CampaignsDotComPage: React.FunctionComponent<CampaignsDotComPageProps> = () => (
    <>
        <CampaignsMarketing
            body={
                <section className="my-3 text-center">
                    <h1>Start using campaigns</h1>

                    <p className="lead">
                        Campaigns are not available on Sourcegraph.com
                        <a href="" rel="noopener">
                            Install Sourcegraph locally
                        </a>{' '}
                        with a free private instance and{' '}
                        <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener">
                            update the configuration settings
                        </a>{' '}
                        to use campaigns.
                    </p>

                    <a href="https://docs.sourcegraph.com/admin/install" rel="noopener" className="btn btn-primary">
                        Get started now
                    </a>
                </section>
            }
        />
    </>
)
