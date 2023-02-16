import React, { useMemo } from 'react'

import { UserAvatar, UserAvatarData } from '@sourcegraph/shared/src/components/UserAvatar'
import { OwnerMatch } from '@sourcegraph/shared/src/search/stream'

import { ResultContainer } from './ResultContainer'

import styles from './OwnerSearchResult.module.scss'
import resultStyles from './SearchResult.module.scss'
import classNames from 'classnames'
import { Link } from '@sourcegraph/wildcard'

export interface PersonSearchResultProps {
    result: OwnerMatch
    onSelect: () => void
    containerClassName?: string
    as?: React.ElementType
    index: number
}

export const OwnerSearchResult: React.FunctionComponent<PersonSearchResultProps> = ({
    result,
    onSelect,
    containerClassName,
    as,
    index,
}) => {
    const displayName = useMemo(() => {
        let displayName = ''
        if (result.type === 'person') {
            displayName = result.displayName || result.username || result.handle || result.email || 'Unknown person'
        } else if (result.type === 'team') {
            displayName = result.displayName || result.name || result.handle || result.email || 'Unknown team'
        } else {
            displayName = result.handle || result.email || 'Unknown owner'
        }
        return displayName
    }, [result])

    const avatarUser = useMemo(() => {
        const avatarUser: UserAvatarData = { username: displayName, avatarURL: null, displayName: displayName }
        if (result.type === 'person') {
            if (result.username) {
                avatarUser.username = result.username
            }
            if (result.avatarURL) {
                avatarUser.avatarURL = result.avatarURL
            }
        }
        return avatarUser
    }, [result, displayName])

    const title = (
        <div className="d-flex align-items-center">
            <UserAvatar user={avatarUser} className={styles.avatar} size={16} />
            {result.type === 'person' && result.username ? (
                <Link to={`/users/${result.username}`} className="text-muted">
                    {displayName}
                </Link>
            ) : (
                <span className="text-muted">{displayName}</span>
            )}
        </div>
    )

    return (
        <ResultContainer
            className={containerClassName}
            as={as}
            index={index}
            title={title}
            detail={null}
            url="#"
            onClick={onSelect}
        >
            <div className={classNames(resultStyles.searchResultMatch, 'p-2 flex-column')}>
                <div className={resultStyles.matchType}>
                    <small>Owner match</small>
                </div>
                {result.type === 'unknownOwner' && (
                    <>
                        <div className={resultStyles.dividerVertical} />
                        <small className="d-block font-italic">
                            This owner is not associated with any Sourcegraph user or team.
                        </small>
                    </>
                )}
            </div>
        </ResultContainer>
    )
}
