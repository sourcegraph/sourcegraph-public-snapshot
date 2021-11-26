import { LocationDescriptor } from 'history'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { Timestamp } from '../../../../components/time/Timestamp'
import {
    ComponentAuthorsFields,
    ComponentUsageFields,
    ComponentStateDetailFields,
    ComponentCodeOwnersFields,
    ComponentStatusFields,
    ComponentOwnerFields,
} from '../../../../graphql-operations'
import { PersonLink } from '../../../../person/PersonLink'
import { ComponentOwner } from '../../components/entity-owner/EntityOwner'

import { OverviewStatusContextItem } from './OverviewStatusContextItem'

interface Props {
    component: ComponentStateDetailFields
    itemClassName?: string
}

export const OverviewStatusContexts: React.FunctionComponent<Props> = ({ component, itemClassName }) => (
    <>
        {component.status.contexts.map(statusContext => {
            switch (statusContext.name) {
                case 'owner':
                    return false ? (
                        <OwnerStatusContext
                            key={statusContext.id}
                            entity={component}
                            statusContext={statusContext}
                            className={itemClassName}
                        />
                    ) : null

                case 'codeOwners':
                    return (
                        <CodeOwnersStatusContext
                            key={statusContext.id}
                            entity={component}
                            statusContext={statusContext}
                            className={itemClassName}
                        />
                    )

                case 'contributors':
                    return (
                        <ContributorsStatusContext
                            key={statusContext.id}
                            entity={component}
                            statusContext={statusContext}
                            className={itemClassName}
                        />
                    )

                case 'usage':
                    return (
                        <UsageStatusContext
                            key={statusContext.id}
                            entity={component}
                            statusContext={statusContext}
                            className={itemClassName}
                        />
                    )

                default:
                    return (
                        <OverviewStatusContextItem
                            key={statusContext.id}
                            statusContext={statusContext}
                            className={itemClassName}
                        />
                    )
            }
        })}
    </>
)

const OwnerStatusContext: React.FunctionComponent<{
    entity: ComponentOwnerFields
    statusContext: ComponentStatusFields['status']['contexts'][0]
    className?: string
}> = ({ entity, statusContext, className }) => (
    <OverviewStatusContextItem statusContext={statusContext} className={className}>
        {!statusContext.description && <ComponentOwner owner={entity.owner} />}
    </OverviewStatusContextItem>
)

const CodeOwnersStatusContext: React.FunctionComponent<{
    entity: ComponentCodeOwnersFields
    statusContext: ComponentStatusFields['status']['contexts'][0]
    className?: string
}> = ({ entity, statusContext, className }) => (
    <OverviewStatusContextItem statusContext={statusContext} className={className}>
        {!statusContext.description && (
            <TruncatedList
                tag="ol"
                className="list-inline mb-0"
                moreUrl={`${statusContext.targetURL!}/code`}
                moreClassName="list-inline-item"
                moreLinkClassName="text-muted small"
            >
                {entity.codeOwners?.map(codeOwner => (
                    <li key={codeOwner.node.email} className="list-inline-item mr-2">
                        {codeOwner.node.email}
                        <span
                            className="small text-muted ml-1"
                            title={`Owns ${codeOwner.fileCount} ${pluralize('file', codeOwner.fileCount)}`}
                        >
                            {codeOwner.fileProportion >= 0.01
                                ? `${(codeOwner.fileProportion * 100).toFixed(0)}%`
                                : '<1%'}
                        </span>
                    </li>
                ))}
            </TruncatedList>
        )}
    </OverviewStatusContextItem>
)

const ContributorsStatusContext: React.FunctionComponent<{
    entity: ComponentAuthorsFields
    statusContext: ComponentStatusFields['status']['contexts'][0]
    className?: string
}> = ({ entity, statusContext, className }) => (
    <OverviewStatusContextItem statusContext={statusContext} className={className}>
        {!statusContext.description && (
            <TruncatedList
                tag="ol"
                className="list-inline mb-0"
                moreUrl={`${statusContext.targetURL!}/code`}
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
        )}
    </OverviewStatusContextItem>
)

const UsageStatusContext: React.FunctionComponent<{
    entity: ComponentUsageFields
    statusContext: ComponentStatusFields['status']['contexts'][0]
    className?: string
}> = ({ entity, statusContext, className }) => (
    <OverviewStatusContextItem statusContext={statusContext} className={className}>
        {!statusContext.description && (
            <TruncatedList
                tag="ol"
                className="list-inline mb-0"
                moreUrl={`${statusContext.targetURL!}/code`}
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
        )}
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
