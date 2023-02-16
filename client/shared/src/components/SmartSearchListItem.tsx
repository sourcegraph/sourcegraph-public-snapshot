import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { formatSearchParameters } from '@sourcegraph/common'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { ProposedQuery } from '@sourcegraph/shared/src/search/stream'
import { Link, createLinkUrl, Icon, Text } from '@sourcegraph/wildcard'

import styles from './SmartSearchPreview.module.scss'

interface SmartSearchListItemProps {
    proposedQuery: ProposedQuery
    previewStyle?: boolean
}

const processDescription = (description: string): string => {
    const split = description.split(' ⚬ ')

    split[0] = split[0][0].toUpperCase() + split[0].slice(1)
    return split.join(', ')
}

export const SmartSearchListItem: React.FunctionComponent<SmartSearchListItemProps> = ({
    proposedQuery,
    previewStyle = false,
}) => (
    <li key={proposedQuery.query} className={classNames(previewStyle ? 'py-2' : styles.listItem)}>
        <Link
            to={createLinkUrl({
                pathname: '/search',
                search: formatSearchParameters(new URLSearchParams({ q: proposedQuery.query })),
            })}
            className={classNames(previewStyle ? 'text-decoration-none' : styles.links)}
        >
            <Text className="mb-0">
                <span className={classNames(previewStyle ? 'text-muted' : styles.listItemDescription)}>
                    {processDescription(proposedQuery.description || '')}
                </span>
                <Icon svgPath={mdiArrowRight} aria-hidden={true} className="mx-2 text-body" />
                <span className={classNames(previewStyle ? 'p-1 bg-code' : styles.suggestion)}>
                    <SyntaxHighlightedSearchQuery
                        query={proposedQuery.query}
                        searchPatternType={SearchPatternType.standard}
                    />
                </span>
                {proposedQuery.annotations
                    ?.filter(({ name }) => name === 'ResultCount')
                    ?.map(({ name, value }) => (
                        <span key={name} className="text-muted ml-2">
                            {' '}
                            ({value})
                        </span>
                    ))}
            </Text>
        </Link>
    </li>
)
