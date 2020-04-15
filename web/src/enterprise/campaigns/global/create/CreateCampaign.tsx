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
        <h1>Create a new campaign</h1>
        <ul className="list-group">
            <li className="list-group-item p-3">
                <Link to="new" className="text-decoration-none">
                    <div className="d-flex">
                        <h3>
                            <CampaignsIcon className="mr-3" />
                        </h3>
                        <div>
                            <h3>Create a new empty campaign</h3>
                            <p className="mb-0">
                                Track existing changesets with Sourcegraph by manually adding them to a campaign. All
                                added changesets can be managed and monitored, while a burndown chart will give you an
                                overview on the progress.
                            </p>
                        </div>
                    </div>
                </Link>
            </li>
            <li className="list-group-item p-3">
                <Link to="cli" className="text-decoration-none">
                    <div className="d-flex">
                        <h3>
                            <CampaignsIcon className="mr-3" />
                        </h3>
                        <div>
                            <h3>Create a campaign from patches</h3>
                            <p className="mb-0">
                                Use the src CLI to make code changes across multiple repositories and turn the resulting
                                set of patches into changesets (pull requests) on code hosts by creating a campaign.
                                Manage and track the progress of the changesets in the newly created campaign.
                            </p>
                        </div>
                    </div>
                </Link>
            </li>
        </ul>
    </div>
)
