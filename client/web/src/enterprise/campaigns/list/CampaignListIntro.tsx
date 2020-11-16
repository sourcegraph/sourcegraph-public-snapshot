import React from 'react'
import { SourcegraphIcon } from '../../../auth/icons'

export const CampaignListIntro: React.FunctionComponent = () => (
    <div className="row">
        <div className="col-12 col-md-6 mb-2">
            <div className="campaign-list-intro__card card p-2 h-100">
                <div className="card-body d-flex align-items-start">
                    {/* d-none d-sm-block ensure that we hide the icon on XS displays. */}
                    <SourcegraphIcon className="mr-3 col-2 mt-2 d-none d-sm-block" />
                    <div>
                        <h4>Campaigns trial</h4>
                        <p className="text-muted mb-0">
                            Campaigns will be a paid feature in a future release. In the meantime, we invite you to
                            trial the ability to make large scale changes across many repositories and code hosts. If
                            you’d like to discuss use cases and features,{' '}
                            <a href="https://about.sourcegraph.com/contact/sales/">please get in touch</a>!
                        </p>
                    </div>
                </div>
            </div>
        </div>
        <div className="col-12 col-md-6 mb-2">
            <div className="campaign-list-intro__card card h-100 p-2">
                <div className="card-body">
                    <h4>New campaigns features in version 3.22</h4>
                    <ul className="text-muted mb-0 pl-3">
                        <li>All users can now create campaigns</li>
                        <li>Publishing a changeset requires a code host token from the user applying the campaign</li>
                        <li>
                            Template variables such as <code>search_result_paths</code> and <code>modified_files</code>{' '}
                            are now available in campaign specifications
                        </li>
                        <li>Rate limit improvements</li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
)
