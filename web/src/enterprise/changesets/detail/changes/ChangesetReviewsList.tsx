import React from 'react'
import { ThreadSettings } from '../../../threads/settings'
import { ChangesetReviewLink } from './ChangesetReviewLink'

interface Props {
    threadSettings: ThreadSettings

    showIcon?: boolean
    className?: string
    linkClassName?: string
}

/**
 * A list of reviewable pull requests in a changeset.
 */
export const ChangesetReviewsList: React.FunctionComponent<Props> = ({
    threadSettings,
    showIcon = true,
    className = '',
    linkClassName = 'btn btn-link border',
}) => (
    <div className={`changeset-reviews-list ${className}`}>
        {threadSettings.relatedPRs && threadSettings.relatedPRs.length > 0 ? (
            <ul className="list-unstyled mb-0 d-flex flex-wrap">
                {threadSettings.relatedPRs.map((link, i) => (
                    <li key={i} className="mr-3 mb-3">
                        <ChangesetReviewLink
                            link={link}
                            showRepositoryName={true}
                            showIcon={showIcon}
                            className={linkClassName}
                        />
                    </li>
                ))}
            </ul>
        ) : (
            <span className="text-muted">No pull requests</span>
        )}
    </div>
)
