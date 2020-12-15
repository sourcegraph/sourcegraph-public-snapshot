import React from 'react'
import { DismissibleAlert } from '../../../components/DismissibleAlert'
import { SourcegraphIcon } from '../../../auth/icons'

export interface CampaignListIntroProps {
    licensed: boolean | undefined
}

export const CampaignListIntro: React.FunctionComponent<CampaignListIntroProps> = ({ licensed }) => (
    <div className="row mb-2">
        {licensed === true ? (
            <div className="col-12">
                <CampaignChangelogAlert />
            </div>
        ) : (
            <>
                {licensed === false && (
                    <>
                        <div className="col-12 col-md-6">
                            <CampaignUnlicensedAlert />
                        </div>
                        <div className="col-12 col-md-6">
                            <CampaignChangelogAlert />
                        </div>
                    </>
                )}
            </>
        )}
    </div>
)

const CampaignChangelogAlert: React.FunctionComponent = () => (
    <DismissibleAlert className="campaign-list-intro__alert" partialStorageKey="campaign-list-intro-changelog-3.23">
        <div className="campaign-list-intro__card card h-100 p-2">
            <div className="card-body">
                <h4>New campaigns features in version 3.23</h4>
                <ul className="text-muted mb-0 pl-3">
                    <li>Changesets in a campaign can be searched by title and repository name</li>
                    <li>
                        Multiple changesets can be created from a single repository using the experimental{' '}
                        <a
                            href="https://docs.sourcegraph.com/campaigns/references/campaign_spec_yaml_reference#transformchanges"
                            target="_blank"
                            rel="noopener"
                            className="text-monospace"
                        >
                            transformChanges
                        </a>{' '}
                        block in a campaign spec
                    </li>
                    <li>
                        Steps in campaign specs can now{' '}
                        <a
                            href="https://docs.sourcegraph.com/campaigns/references/campaign_spec_yaml_reference#examples-7"
                            target="_blank"
                            rel="noopener"
                        >
                            pass environment variables
                        </a>{' '}
                        on to containers
                    </li>
                    <li>
                        The preview of a campaign spec now includes detailed information about what will change, instead
                        of only showing the desired state
                    </li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)

const CampaignUnlicensedAlert: React.FunctionComponent = () => (
    <div className="campaign-list-intro__alert">
        <div className="campaign-list-intro__card card p-2 h-100">
            <div className="card-body d-flex align-items-start">
                {/* d-none d-sm-block ensure that we hide the icon on XS displays. */}
                <SourcegraphIcon className="mr-3 col-2 mt-2 d-none d-sm-block" />
                <div>
                    <h4>Campaigns trial</h4>
                    <p className="text-muted">
                        Campaigns is a paid feature of Sourcegraph. All users can create sample campaigns with up to
                        five changesets without a license.
                    </p>
                    <p className="text-muted mb-0">
                        <a href="https://about.sourcegraph.com/contact/sales/">Contact sales</a> to obtain a trial
                        license.
                    </p>
                </div>
            </div>
        </div>
    </div>
)
