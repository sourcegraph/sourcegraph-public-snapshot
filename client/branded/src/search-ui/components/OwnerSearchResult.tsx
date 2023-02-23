import React, { useMemo } from 'react'

import classNames from 'classnames'

import { UserAvatar, UserAvatarData } from '@sourcegraph/shared/src/components/UserAvatar'
import { getOwnerMatchUrl, OwnerMatch } from '@sourcegraph/shared/src/search/stream'
import { Link } from '@sourcegraph/wildcard'

import { ResultContainer } from './ResultContainer'

import styles from './OwnerSearchResult.module.scss'
import resultStyles from './SearchResult.module.scss'

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
            displayName =
                result.user?.displayName || result.user?.username || result.handle || result.email || 'Unknown person'
        } else if (result.type === 'team') {
            displayName = result.displayName || result.name || result.handle || result.email || 'Unknown team'
        }
        return displayName
    }, [result])

    const avatarUser = useMemo(() => {
        const avatarUser: UserAvatarData = { username: displayName, avatarURL: null, displayName }
        if (result.type === 'person' && result.user) {
            avatarUser.username = result.user.username
            avatarUser.avatarURL = result.user.avatarURL || null
        }
        return avatarUser
    }, [result, displayName])

    const url = useMemo(() => {
        const url = getOwnerMatchUrl(result)
        if (result.type === 'person' && !result.user) {
            // This is not a real URL, remove it.
            return ''
        }
        return url
    }, [result])

    const title = (
        <div className="d-flex align-items-center">
            <UserAvatar user={avatarUser} className={styles.avatar} size={16} />
            {url ? (
                <Link to={url} className="text-muted">
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
                {result.type === 'person' && !result.user && (
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
