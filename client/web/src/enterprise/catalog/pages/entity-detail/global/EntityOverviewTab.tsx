import classNames from 'classnames'
import { LocationDescriptor } from 'history'
import { uniqBy } from 'lodash'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { CatalogEntityRelationType } from '@sourcegraph/shared/src/graphql/schema'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { Timestamp } from '../../../../../components/time/Timestamp'
import {
    CatalogComponentAuthorsFields,
    CatalogComponentUsageFields,
    CatalogEntityDetailFields,
    CatalogEntityOwnersFields,
} from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { EntityGraph } from '../../../components/entity-graph/EntityGraph'

import { ComponentSourceDefinitions } from './ComponentSourceDefinitions'
import { EntityDetailContentCardProps } from './EntityDetailContent'

interface Props extends EntityDetailContentCardProps {
    entity: CatalogEntityDetailFields
}

export const EntityOverviewTab: React.FunctionComponent<Props> = ({ entity, className }) => {
    const searchSourcesCard = (
        <div className="card mb-3">
            <div className="card-body">
                <Link
                    to={`/search?q=context:c/${entity.name}`}
                    className="d-flex align-items-center btn btn-outline-secondary"
                >
                    <SearchIcon className="icon-inline" /> Search in {entity.name}...
                </Link>
            </div>
            <ComponentSourceDefinitions
                catalogComponent={entity}
                listGroupClassName="list-group-flush"
                className="border-top"
            />
        </div>
    )

    return (
        <div className={classNames('d-flex flex-column', className)}>
            {entity.__typename === 'CatalogComponent' ? (
                <>
                    <div className="row">
                        <div className="col-md-8">
                            {searchSourcesCard}
                            {false && entity.commits?.nodes[0] && (
                                <LastCommit commit={entity.commits.nodes[0]} className="" />
                            )}
                            <OwnersInfoBox entity={entity} className="mb-3" />
                            <AuthorsInfoBox entity={entity} className="mb-3" />
                            <UsageInfoBox entity={entity} className="mb-3" />
                        </div>
                        <div className="col-md-4">
                            {/* owner-docs-API def -- authorities. then who you could ask. */}
                            {entity.description && <p className="mb-3">{entity.description}</p>}
                            <div>
                                <Link to="#" className="d-flex align-items-center text-body mb-3 mr-3">
                                    <FileDocumentIcon className="icon-inline mr-2" />
                                    Documentation
                                </Link>
                                <Link to="#" className="d-flex align-items-center text-body mb-3 mr-3">
                                    <FileAlertIcon className="icon-inline mr-2" />
                                    Runbook
                                </Link>
                                <Link to="#" className="d-flex align-items-center text-body mb-3">
                                    <AlertCircleOutlineIcon className="icon-inline mr-2" />
                                    Issues
                                </Link>
                                <Link to="#" className="d-flex align-items-center text-body mb-3">
                                    <SlackIcon className="icon-inline mr-2" />
                                    #dev-frontend
                                </Link>
                            </div>
                        </div>
                    </div>
                    <EntityGraph
                        graph={{
                            edges: entity.relatedEntities.edges.map(edge =>
                                edge.type === CatalogEntityRelationType.DEPENDS_ON
                                    ? {
                                          type: edge.type,
                                          outNode: entity,
                                          inNode: edge.node,
                                      }
                                    : {
                                          type: CatalogEntityRelationType.DEPENDS_ON,
                                          outNode: edge.node,
                                          inNode: entity,
                                      }
                            ),
                            nodes: uniqBy(entity.relatedEntities.edges.map(edge => edge.node).concat(entity), 'id'),
                        }}
                        activeNodeID={entity.id}
                        className="border-top my-3"
                    />
                </>
            ) : (
                <div>Typename is {entity.__typename}</div>
            )}
        </div>
    )
}

const LastCommit: React.FunctionComponent<{
    commit: NonNullable<CatalogEntityDetailFields['commits']>['nodes'][0]
    className?: string
}> = ({ commit, className }) => (
    <div className={classNames('d-flex align-items-center', className)}>
        <UserAvatar className="icon-inline mr-2 flex-shrink-0" user={commit.author.person} size={12} />
        <PersonLink person={commit.author.person} className="font-weight-bold mr-2 flex-shrink-0" />
        <Link to={commit.url} className="text-truncate flex-grow-1 text-body mr-2" title={commit.message}>
            {commit.subject}
        </Link>
        <small className="text-nowrap text-muted">
            <Link to={commit.url} className="text-monospace text-muted mr-2 d-none d-md-inline">
                {commit.abbreviatedOID}
            </Link>
            <Timestamp date={commit.author.date} noAbout={true} />
        </small>
    </div>
)

const InfoBox: React.FunctionComponent<{
    title: string
    titleUrl?: LocationDescriptor
    color: 'success' | 'primary' | 'warning' | 'danger'
    className?: string
}> = ({ title, titleUrl, color, className, children }) => (
    <div className={classNames('d-flex align-items-start', className)}>
        <h4
            className={classNames(`badge bg-transparent mb-0 mr-2 border border-${color} text-${color}`)}
            // eslint-disable-next-line react/forbid-dom-props
            style={{ marginTop: '-1px' }}
        >
            {titleUrl ? (
                <Link to={titleUrl} className={`text-${color}`}>
                    {title}
                </Link>
            ) : (
                title
            )}
        </h4>
        <div>{children}</div>
    </div>
)

const OwnersInfoBox: React.FunctionComponent<{
    entity: CatalogEntityOwnersFields & Pick<CatalogEntityDetailFields, 'url'>
    className?: string
}> = ({ entity, className }) => (
    <InfoBox title="Owners" titleUrl={`${entity.url}/code`} color="success" className={className}>
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
    </InfoBox>
)

const AuthorsInfoBox: React.FunctionComponent<{
    entity: CatalogComponentAuthorsFields & Pick<CatalogEntityDetailFields, 'url'>
    className?: string
}> = ({ entity, className }) => (
    <InfoBox title="Authors" titleUrl={`${entity.url}/code`} color="primary" className={className}>
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
    </InfoBox>
)

const UsageInfoBox: React.FunctionComponent<{
    entity: CatalogComponentUsageFields & Pick<CatalogEntityDetailFields, 'url'>
    className?: string
}> = ({ entity, className }) => (
    <InfoBox title="Usage" titleUrl={`${entity.url}/usage`} color="primary" className={className}>
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
    </InfoBox>
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
