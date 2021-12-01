import { LocationDescriptor } from 'history'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { CatalogEntityStatusState } from '@sourcegraph/shared/src/graphql/schema'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { Timestamp } from '../../../../../components/time/Timestamp'
import {
    CatalogComponentAuthorsFields,
    CatalogComponentUsageFields,
    CatalogEntityDetailFields,
    CatalogEntityOwnersFields,
} from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'

import { OverviewStatusContextItem } from './OverviewStatusContextItem'

export const OwnersStatusContext: React.FunctionComponent<{
    entity: CatalogEntityOwnersFields & Pick<CatalogEntityDetailFields, 'url'>
    className?: string
}> = ({ entity, className }) => (
    <OverviewStatusContextItem
        statusContext={{
            name: 'owners',
            title: 'Owners',
            description: '',
            state: entity.owners?.length > 0 ? CatalogEntityStatusState.INFO : CatalogEntityStatusState.FAILURE,
            targetURL: `${entity.url}/code`,
        }}
        className={className}
    >
        {entity.owners?.length > 0 ? (
            <TruncatedList
                tag="ol"
                className="list-inline mb-0"
                moreUrl={`${entity.url}/code`}
                moreClassName="list-inline-item"
                moreLinkClassName="text-muted small"
            >
                {entity.owners?.map(owner => (
                    <li key={owner.node} className="list-inline-item mr-2">
                        {owner.node}
                        <span
                            className="small text-muted ml-1"
                            title={`Owns ${owner.fileCount} ${pluralize('file', owner.fileCount)}`}
                        >
                            {owner.fileProportion >= 0.01 ? `${(owner.fileProportion * 100).toFixed(0)}%` : '<1%'}
                        </span>
                    </li>
                ))}
            </TruncatedList>
        ) : (
            <span>No code owners found</span>
        )}
    </OverviewStatusContextItem>
)

export const AuthorsStatusContext: React.FunctionComponent<{
    entity: CatalogComponentAuthorsFields & Pick<CatalogEntityDetailFields, 'url'>
    className?: string
}> = ({ entity, className }) => (
    <OverviewStatusContextItem
        statusContext={{
            name: 'authors',
            title: 'Authors',
            description: '',
            state: CatalogEntityStatusState.INFO,
            targetURL: `${entity.url}/code`,
        }}
        className={className}
    >
        <TruncatedList
            tag="ol"
            className="list-inline mb-0"
            moreUrl={`${entity.url}/code`}
            moreClassName="list-inline-item"
            moreLinkClassName="text-muted small"
        >
            {entity.authors?.map(author => (
                <li key={author.person.email} className="list-inline-item mr-2">
                    <PersonLink person={author.person} />
                    <span
                        className="small text-muted ml-1"
                        title={`${author.authoredLineCount} ${pluralize('line', author.authoredLineCount)}`}
                    >
                        {author.authoredLineProportion >= 0.01
                            ? `${(author.authoredLineProportion * 100).toFixed(0)}%`
                            : '<1%'}
                    </span>
                    <span className="small text-muted ml-1">
                        <Timestamp date={author.lastCommit.author.date} noAbout={true} />
                    </span>
                </li>
            ))}
        </TruncatedList>
    </OverviewStatusContextItem>
)

export const UsageStatusContext: React.FunctionComponent<{
    entity: CatalogComponentUsageFields & Pick<CatalogEntityDetailFields, 'url'>
    className?: string
}> = ({ entity, className }) => (
    <OverviewStatusContextItem
        statusContext={{
            name: 'usage',
            title: 'Usage',
            description: '',
            state: CatalogEntityStatusState.INFO,
            targetURL: `${entity.url}/usage`,
        }}
        className={className}
    >
        <TruncatedList
            tag="ol"
            className="list-inline mb-0"
            moreUrl={`${entity.url}/code`}
            moreClassName="list-inline-item"
            moreLinkClassName="text-muted small"
        >
            {entity.usage?.people.map(edge => (
                <li key={edge.node.email} className="list-inline-item mr-2">
                    <PersonLink person={edge.node} />
                    <span className="small text-muted ml-1">
                        {edge.authoredLineCount} {pluralize('use', edge.authoredLineCount)}
                    </span>
                    <span className="small text-muted ml-1">
                        <Timestamp date={edge.lastCommit.author.date} noAbout={true} />
                    </span>
                </li>
            ))}
        </TruncatedList>
    </OverviewStatusContextItem>
)

const useListSeeMore = <T extends any>(list: T[], max: number): [T[], boolean] => {
    if (list.length > max) {
        return [list.slice(0, max), true]
    }
    return [list, false]
}

const TruncatedList: React.FunctionComponent<{
    tag: 'ol' | 'ul'
    max?: number
    className?: string
    moreUrl?: LocationDescriptor
    moreClassName?: string
    moreLinkClassName?: string
}> = ({ tag: Tag, children, max = 5, className, moreUrl, moreClassName, moreLinkClassName }) => {
    const childrenArray = React.Children.toArray(children)
    const [firstChildren, seeMore] = useListSeeMore(childrenArray, max)
    return (
        <Tag className={className}>
            {firstChildren}
            {seeMore && (
                <li className={moreClassName}>
                    <LinkOrSpan to={moreUrl} className={moreLinkClassName}>
                        ...{childrenArray.length - max} more
                    </LinkOrSpan>
                </li>
            )}
        </Tag>
    )
}
