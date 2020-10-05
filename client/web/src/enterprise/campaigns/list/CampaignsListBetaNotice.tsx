import React from 'react'

export const CampaignsListBetaNotice: React.FunctionComponent<{}> = () => (
    <div className="row d-none">
        <div className="col-12 col-md-8 offset-md-2">
            <div className="card bg-white mt-4 mb-4">
                <div className="card-body p-3 d-flex">
                    <img
                        className="p-3 mr-3 campaigns-list-beta-notice__logo"
                        src="/.assets/img/sourcegraph-mark.svg"
                    />
                    <div>
                        <h3>Welcome to campaigns beta!</h3>
                        <p className="mb-1">
                            We're excited for you to use campaigns to remove legacy code, fix critical security issues,
                            pay down tech debt, and more. We look forward to hearing about campaigns you run inside your
                            organization. See{' '}
                            <a href="https://docs.sourcegraph.com/user/campaigns">campaigns documentation</a>, and{' '}
                            <a href="mailto:feedback@sourcegraph.com?subject=Campaigns feedback">get in touch</a> with
                            any questions or feedback!
                        </p>
                    </div>
                </div>
            </div>
        </div>
    </div>
)
