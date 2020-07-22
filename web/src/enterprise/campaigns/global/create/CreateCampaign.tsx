import * as React from 'react'
import { CampaignsIcon } from '../../icons'
import { Link } from '../../../../../../shared/src/components/Link'

interface Props {
    className?: string
}

/**
 * Overview page over the different types of campaigns
 */
export const CreateCampaign: React.FunctionComponent<Props> = ({ className }) => (
    <div className={className}>
        <h1>
            Create a new campaign{' '}
            <sup>
                <span className="badge badge-info">Beta</span>
            </sup>
        </h1>
        <ul className="list-group">
            <li className="list-group-item p-3">
                <Link to="cli" className="text-decoration-none">
                    <div className="d-flex">
                        <h3>
                            <CampaignsIcon className="mr-3" />
                        </h3>
                        <div>
                            <h3>Create and track changesets</h3>
                            <p className="mb-0">
                                Change code in multiple repositories. Turn the resulting set of patches into changesets
                                (pull requests) on code hosts by creating a campaign. Track the progress of the
                                changesets in the newly created campaign.
                            </p>
                        </div>
                    </div>
                </Link>
            </li>
            <li className="list-group-item p-3">
                <Link to="new" className="text-decoration-none">
                    <div className="d-flex">
                        <h3>
                            <CampaignsIcon className="mr-3" />
                        </h3>
                        <div>
                            <h3>Track existing changesets</h3>
                            <p className="mb-0">
                                Track a collection of already created changesets by collecting them in a campaign. The
                                burndown chart provides an overview of progress, and filters help surface which
                                changesets need action.
                            </p>
                        </div>
                    </div>
                </Link>
            </li>
        </ul>
    </div>
)
