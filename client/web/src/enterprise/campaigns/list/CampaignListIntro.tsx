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
    <DismissibleAlert className="campaign-list-intro__alert" partialStorageKey="campaign-list-intro-changelog-3.24">
        <div className="campaign-list-intro__card card h-100 p-2">
            <div className="card-body">
                <h4>New campaigns features in version 3.24</h4>
                <ul className="text-muted mb-0 pl-3">
                    <li>
                        <code>src</code> now executes campaigns{' '}
                        <a href="https://github.com/sourcegraph/src-cli/pull/412">significantly faster</a> on Intel
                        macOS
                    </li>
                    <li>
                        Campaign specs now allow{' '}
                        <a href="https://docs.sourcegraph.com/campaigns/references/campaign_spec_yaml_reference#steps-outputs">
                            step outputs
                        </a>{' '}
                        to be captured for later use in templated fields
                    </li>
                    <li>
                        Campaign spec <code>changesetTemplate</code> fields now support{' '}
                        <a href="https://docs.sourcegraph.com/campaigns/references/campaign_spec_templating">
                            templating
                        </a>
                    </li>
                    <li>
                        When creating or updating a campaign, the preview now provides more information about the
                        operations to be performed on each changeset
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
