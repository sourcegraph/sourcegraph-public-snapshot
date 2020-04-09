import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'
import { Link } from '../../../../../../shared/src/components/Link'

export interface CampaignsSiteAdminMarketingPageProps {}

export const CampaignsSiteAdminMarketingPage: React.FunctionComponent<CampaignsSiteAdminMarketingPageProps> = () => (
    <>
        <CampaignsMarketing />

        <section className="my-3 text-center">
            <h1>Start using campaigns today.</h1>

            <p className="lead">
                Go to Settings and enable Campaigns in the experimentalFeatures section.{' '}
                <code>"automation": "enabled"</code>
            </p>

            <Link to="/site-admin/configuration" className="btn btn-primary">
                Go to my instance settings
            </Link>
        </section>
    </>
)
