import React from 'react'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { OwnerMatch } from '@sourcegraph/shared/src/search/stream'

import { ResultContainer } from './ResultContainer'

import styles from './OwnerSearchResult.module.scss'
import resultStyles from './SearchResult.module.scss'
import classNames from 'classnames'

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
    const title = (
        <div className="d-flex align-items-center">
            <UserAvatar
                user={{ username: result.handle, avatarURL: null, displayName: null }}
                className={styles.avatar}
                size={16}
            />
            {result.handle}
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
