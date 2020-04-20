import React from 'react'
import { CampaignsMarketing } from './CampaignsMarketing'
import { Link } from '../../../../../../shared/src/components/Link'

export interface CampaignsSiteAdminMarketingPageProps {}

export const CampaignsSiteAdminMarketingPage: React.FunctionComponent<CampaignsSiteAdminMarketingPageProps> = () => (
    <CampaignsMarketing
        body={
            <section className="my-3">
                <h2>Get started</h2>

                <p>Creating and managing campaigns is only available to Sourcegraph admins at this time.</p>
                <ol>
                    <li>
                        <a href="https://docs.sourcegraph.com/user/campaigns">
                            Update your site configuration settings
                        </a>{' '}
                        to enable campaigns for site admins.
                        <div className="alert alert-info mt-3">
                            <p>
                                <strong>
                                    By enabling this feature, you agree to test out functionality that is currently in
                                    beta, and acknowledge the following:
                                </strong>
                            </p>
                            <ul className="mb-3">
                                <li>
                                    During the beta period, campaigns are free to use. After the beta period ends,
                                    campaigns will be available as a paid add-on.
                                </li>
                                <li>
                                    Campaigns are not included as part of the Enterprise tiers (Enterprise, Enterprise
                                    Starter, Enterprise Plus, etc.) license tiers.
                                </li>
                            </ul>
                            Please <a href="mailto:sales@sourcegraph.com">contact us</a> for more information.
                        </div>
                    </li>
                    <li>
                        Optional: enable read-only access for all users.
                        <div className="alert alert-warning mt-3">
                            <strong>WARNING:</strong> Repository permissions are NOT enforced if you enable read-only
                            views on campaigns. Any authenticated user will be able to see all code changes associated
                            with a campaign. Therefore, read-only access is recommended for instances without repository
                            permissions configured.
                        </div>
                    </li>
                    <li>
                        <a href="https://docs.sourcegraph.com/user/campaigns">Create your first campaign</a>
                    </li>
                </ol>
                <div>
                    <Link to="/site-admin/configuration" className="btn btn-primary mr-2">
                        Go to my site configuration
                    </Link>
                    <a href="https://docs.sourcegraph.com/user/campaigns" rel="noopener" className="btn btn-primary">
                        Learn how to run campaigns
                    </a>
                </div>
            </section>
        }
    />
)
