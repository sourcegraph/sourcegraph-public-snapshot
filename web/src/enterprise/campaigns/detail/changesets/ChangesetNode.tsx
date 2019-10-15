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
    node: IChangeset
}

export const ChangesetNode: React.FunctionComponent<Props> = ({ node }) => {
    const ReviewStateIcon = changesetReviewStateIcons[node.reviewState]
    return (
        <li className="list-group-item d-flex pl-1 align-items-center">
            <div className="flex-shrink-0 flex-grow-0 m-1">
                <SourcePullIcon
                    className={`text-${changesetStatusColorClasses[node.state]}`}
                    data-tooltip={changesetStageLabels[node.state]}
                />
            </div>
            <div className="flex-shrink-0 flex-grow-0 m-1">
                <ReviewStateIcon
                    className={`text-${changesetReviewStateColors[node.reviewState]}`}
                    data-tooltip={changesetStageLabels[node.reviewState]}
                />
            </div>
            <div className="flex-fill overflow-hidden m-1">
                <h4 className="m-0">
                    <Link to={node.repository.url} className="text-muted">
                        {node.repository.name}
                    </Link>{' '}
                    <Link to={node.externalURL.url} target="_blank" rel="noopener noreferrer">
                        {node.title}
                    </Link>
                </h4>
                <div className="text-truncate w-100">{node.body}</div>
            </div>
        </li>
    )
}
