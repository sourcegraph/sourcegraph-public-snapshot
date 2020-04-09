import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'

export interface CampaignsDotComPageProps {}

export const CampaignsDotComPage: React.FunctionComponent<CampaignsDotComPageProps> = () => (
    <>
        <CampaignsMarketing />

        <section className="my-3 text-center">
            <h1>Start using campaigns today.</h1>

            <p className="lead">
                To use campaigns, a private Sourcegraph instance is required. If your organization has one set up
                already, just navigate to Campaigns from the top navigation.
            </p>

            <a href="https://about.sourcegraph.com" rel="noopener" className="btn btn-primary">
                Get started with my instance
            </a>
        </section>
    </>
)
