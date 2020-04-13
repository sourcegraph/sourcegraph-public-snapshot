import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'
import { Link } from '../../../../../../shared/src/components/Link'

export interface CampaignsSiteAdminMarketingPageProps {}

export const CampaignsSiteAdminMarketingPage: React.FunctionComponent<CampaignsSiteAdminMarketingPageProps> = () => (
    <>
        <CampaignsMarketing
            body={
                <section className="my-3 text-center">
                    <h1>Enable campaigns for your team</h1>

                    <p className="lead">
                        <a href="https://docs.sourcegraph.com/user/campaigns">Update your configuration settings</a> to
                        make large-scale code changes across many repositories and different code hosts.
                    </p>
                    <div className="alert alert-info text-left">
                        <h4>
                            By enabling this feature, you are agreeing to test out functionality that is currently in
                            beta.
                        </h4>
                        <ul>
                            <li>
                                This add-on is not part of the Enterprise, Enterprise plus license tiers and at some
                                point in the future will need to be negotiated.
                            </li>
                            <li>
                                You will currently incur no charges for this, but in the future you will be charged for
                                this as an add-on. Please <a href="mailto:sales@sourcegraph.com">contact us</a> to
                                upgrade to this feature.
                            </li>
                            <li>
                                <strong>Enabling read-only for all users</strong>
                                <br />
                                <strong>Warning:</strong> There are no repository permissions enforced if you enable
                                read-only views on campaigns. Any logged in user will be able to see all code changes
                                associated with a campaign.
                            </li>
                        </ul>
                    </div>
                    <div>
                        <Link to="/site-admin/configuration" className="btn btn-primary mr-2">
                            Go to my instance settings
                        </Link>
                        <a
                            href="https://docs.sourcegraph.com/user/campaigns"
                            rel="noopener"
                            className="btn btn-primary"
                        >
                            Getting started with campaigns
                        </a>
                    </div>
                </section>
            }
        />
    </>
)
