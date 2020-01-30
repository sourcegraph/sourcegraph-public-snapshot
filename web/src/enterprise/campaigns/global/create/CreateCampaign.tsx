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
        <h1>
            Create a new campaign
        </h1>
        <ul className="list-group">
            <li className="list-group-item p-3">
                <Link to="new" className="text-decoration-none">
                    <div className="d-flex">
                        <h3>
                            <CampaignsIcon className="mr-3" />
                        </h3>
                        <div>
                            <h3>Manual campaign</h3>
                            <p className="mb-0">
                                Choose manual campaign when you want to track existing changesets with Sourcegraph. All
                                added changesets can be tracked for their merge status and a burndown chart will give
                                you an overview on the progress.
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
                            <h3>Automatic campaign</h3>
                            <p className="mb-0">
                                Automatic campaigns are a powerful way to perform large-scale code changes through
                                Sourcegraph. The `src` cli will allow you to generate all those desired changes. Then,
                                go by the Sourcegraph UI, get an overview of all the changes and gradually or
                                all-at-once roll out your code changes to hundreds of repositories across different code
                                hosts.
                            </p>
                        </div>
                    </div>
                </Link>
            </li>
        </ul>
    </div>
)
