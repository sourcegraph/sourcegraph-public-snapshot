import H from 'history'
import AlertOutlineIcon from 'mdi-react/AlertOutlineIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import * as React from 'react'
import { CHECKS } from '../data'
import { ChecksListHeader } from './ChecksListHeader'
import { ChecksListHeaderFilterButtonDropdown } from './ChecksListHeaderFilterButtonDropdown'
import { ChecksListItem } from './ChecksListItem'

interface Props {
    location: H.Location
}

/**
 * The list of checks with a header.
 */
export const ChecksList: React.FunctionComponent<Props> = ({ location }) => (
    <div className="checks-list">
        <ChecksListHeader location={location} />
        <div className="card">
            <div className="card-header d-flex align-items-center justify-content-between">
                <div className="form-check mx-2">
                    <input className="form-check-input position-static" type="checkbox" aria-label="Select item" />
                </div>
                <div className="font-weight-normal flex-1">
                    <strong>
                        <AlertOutlineIcon className="icon-inline" /> 8 open &nbsp;{' '}
                    </strong>
                    <CheckIcon className="icon-inline" /> 27 closed
                </div>
                <div>
                    <ChecksListHeaderFilterButtonDropdown
                        header="Filter by who's assigned"
                        items={['sqs (you)', 'ekonev', 'jleiner', 'ziyang', 'kting7', 'ffranksena']}
                    >
                        Assignee
                    </ChecksListHeaderFilterButtonDropdown>
                    <ChecksListHeaderFilterButtonDropdown
                        header="Filter by label"
                        items={[
                            'perf',
                            'tech-lead',
                            'services',
                            'bugs',
                            'build',
                            'noisy',
                            'security',
                            'appsec',
                            'infosec',
                            'compliance',
                            'docs',
                        ]}
                    >
                        Labels
                    </ChecksListHeaderFilterButtonDropdown>
                    <ChecksListHeaderFilterButtonDropdown
                        header="Sort by"
                        items={['Priority', 'Most recently updated', 'Least recently updated']}
                    >
                        Sort
                    </ChecksListHeaderFilterButtonDropdown>
                </div>
            </div>
            <ul className="list-group list-group-flush">
                {CHECKS.map((check, i) => (
                    <ChecksListItem key={i} location={location} check={check} />
                ))}
            </ul>
        </div>
    </div>
)
