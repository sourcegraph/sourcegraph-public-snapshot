import { IChangeset } from '../../../../../../shared/src/graphql/schema'
import React from 'react'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import {
    changesetStatusColorClasses,
    changesetReviewStateColors,
    changesetReviewStateIcons,
    changesetStageLabels,
} from './presentation'
import { Link } from '../../../../../../shared/src/components/Link'

interface Props {
    changesets: IChangeset[]
}

export const ChangesetList: React.FunctionComponent<Props> = ({ changesets }) => (
    <ul className="list-group">
        {changesets.map(changeset => {
            const ReviewStateIcon = changesetReviewStateIcons[changeset.reviewState]
            return (
                <li key={changeset.id} className="list-group-item d-flex pl-1 align-items-center">
                    <div className="flex-shrink-0 flex-grow-0 m-1">
                        <SourcePullIcon
                            className={`text-${changesetStatusColorClasses[changeset.state]}`}
                            data-tooltip={changesetStageLabels[changeset.state]}
                        />
                    </div>
                    <div className="flex-shrink-0 flex-grow-0 m-1">
                        <ReviewStateIcon
                            className={`text-${changesetReviewStateColors[changeset.reviewState]}`}
                            data-tooltip={changesetStageLabels[changeset.reviewState]}
                        />
                    </div>
                    <div className="flex-fill overflow-hidden m-1">
                        <h4 className="m-0">
                            <Link to={changeset.repository.url} className="text-muted">
                                {changeset.repository.name}
                            </Link>{' '}
                            <Link to={changeset.externalURL.url} target="_blank" rel="noopener noreferrer">
                                {changeset.title}
                            </Link>
                        </h4>
                        <div className="text-truncate w-100">{changeset.body}</div>
                    </div>
                </li>
            )
        })}
    </ul>
)
