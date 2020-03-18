import * as React from 'react'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { CampaignsIcon } from '../../icons'
import { Link } from '../../../../../../shared/src/components/Link'

interface Props extends ThemeProps {
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
                            <h3>Create a manual campaign</h3>
                            <p className="mb-0">
                                Choose manual campaign if you want to track existing changesets (pull requests) with
                                Sourcegraph. All added changesets can be monitored and a burndown chart will give you an
                                overview on the progress.
                            </p>
                        </div>
                    </div>
                </Link>
            </li>
            <li className="list-group-item p-3">
                <Link to="automated" className="text-decoration-none">
                    <div className="d-flex">
                        <h3>
                            <CampaignsIcon className="mr-3" />
                        </h3>
                        <div>
                            <h3>Create a campaign using the src CLI</h3>
                            <p className="mb-0">
                                When a Campaign is created from a set of patches, one per repository, Sourcegraph will
                                create changesets (pull requests) on the associated code hosts and track their progress
                                in the newly created campaign, where you can manage them.
                                <br />
                                With the src CLI tool, you can not only create the campaign from an existing set of
                                patches, but you can also generate the patches for a number of repositories.
                            </p>
                        </div>
                    </div>
                </Link>
            </li>
        </ul>
    </div>
)
