import type { FC } from 'react'

import classNames from 'classnames'

import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Link } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../../graphql-operations'

import { useDynamicQuery } from './useDynamicQuery'

interface SearchTaskProps {
    label: string
    template: string
    snippets?: string[] | Record<string, string[]>
    handleLinkClick: (event: React.MouseEvent<HTMLElement, MouseEvent> | React.KeyboardEvent<HTMLElement>) => void
}

export const SearchTask: FC<SearchTaskProps> = ({ template, snippets, label, handleLinkClick }) => {
    const selectedQuery = useDynamicQuery(template, snippets)

    return selectedQuery ? (
        <Link
            className={classNames('flex-grow-1')}
            to={`/search?${buildSearchURLQuery(selectedQuery, SearchPatternType.standard, false)}`}
            onClick={handleLinkClick}
        >
            {label}
        </Link>
    ) : null
}
