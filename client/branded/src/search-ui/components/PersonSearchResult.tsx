import { PersonMatch } from '@sourcegraph/shared/src/search/stream'
import React from 'react'
import { ResultContainer } from './ResultContainer'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'

export interface PersonSearchResultProps {
    result: PersonMatch
    onSelect: () => void
    containerClassName?: string
    as?: React.ElementType
    index: number
}

export const PersonSearchResult: React.FunctionComponent<PersonSearchResultProps> = ({
    result,
    onSelect,
    containerClassName,
    as,
    index,
}) => {
    const content = (
        <div className="p-2">
            <UserAvatar user={{ username: result.handle, avatarURL: null, displayName: null }} className="mr-2" />
            {result.handle}
        </div>
    )

    return (
        <ResultContainer
            className={containerClassName}
            as={as}
            index={index}
            title={null}
            detail={null}
            url="#"
            onClick={onSelect}
        >
            {content}
        </ResultContainer>
    )
}
