import classNames from 'classnames'
import { uniqBy } from 'lodash'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import PowerCycleIcon from 'mdi-react/PowerCycleIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React, { useRef, useState } from 'react'
import { Link } from 'react-router-dom'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { CatalogEntityRelationType } from '@sourcegraph/shared/src/graphql/schema'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { CatalogComponentDocumentationFields, CatalogEntityDetailFields } from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { Popover } from '../../../../insights/components/popover/Popover'
import { EntityGraph } from '../../../components/entity-graph/EntityGraph'

import { ComponentSourceDefinitions } from './ComponentSourceDefinitions'
import { EntityDetailContentCardProps } from './EntityDetailContent'
import { OverviewStatusContexts } from './OverviewStatusContexts'

interface Props extends EntityDetailContentCardProps {
    entity: CatalogEntityDetailFields
}

export const EntityOverviewTab: React.FunctionComponent<Props> = ({ entity, className }) => (
    <div className={classNames('d-flex flex-column', className)}>
        {entity.__typename === 'CatalogComponent' ? (
            <>
                <div className="row">
                    <div className="col-md-8">
                        <div className="card mb-3">
                            <ComponentSourceDefinitions
                                catalogComponent={entity}
                                listGroupClassName="list-group-flush"
                            />
                            {entity.commits?.nodes[0] && (
                                <LastCommit commit={entity.commits?.nodes[0]} className="card-footer" />
                            )}
                        </div>
                        <OverviewStatusContexts entity={entity} itemClassName="mb-3" />
                    </div>
                    <div className="col-md-4">
                        {/* owner-docs-API def -- authorities. then who you could ask. */}
                        {entity.description && <p className="mb-3">{entity.description}</p>}
                        <div>
                            <Link
                                to={`/search?q=context:c/${entity.name}`}
                                className="d-inline-flex align-items-center btn btn-outline-secondary mb-3"
                            >
                                <SearchIcon className="icon-inline" /> Search...
                            </Link>
                            {entity.readme && (
                                <div className="d-flex align-items-start">
                                    <Link
                                        to={entity.readme.url}
                                        className="d-flex align-items-center text-body mb-3 mr-2"
                                    >
                                        <FileDocumentIcon className="icon-inline mr-2" />
                                        Documentation
                                    </Link>
                                    <FilePeekButton file={entity.readme} />
                                </div>
                            )}
                            <Link to="#" className="d-flex align-items-center text-body mb-3">
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
                            <hr className="my-3" />
                            <Link to="#" className="d-flex align-items-center text-body mb-3">
                                <PowerCycleIcon className="icon-inline mr-2" />
                                Lifecycle:&nbsp;<strong>{entity.lifecycle.toLowerCase()}</strong>
                            </Link>
                            <Link to="#" className="d-flex align-items-center text-body mb-3">
                                <SettingsIcon className="icon-inline mr-2" />
                                Spec
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

const LastCommit: React.FunctionComponent<{
    commit: NonNullable<CatalogEntityDetailFields['commits']>['nodes'][0]
    className?: string
}> = ({ commit, className }) => (
    <div className={classNames('d-flex align-items-center', className)}>
        <UserAvatar className="icon-inline mr-2 flex-shrink-0" user={commit.author.person} size={18} />
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

const FilePeekButton: React.FunctionComponent<{
    file: NonNullable<CatalogComponentDocumentationFields['readme']>
}> = ({ file }) => {
    const targetButtonReference = useRef<HTMLButtonElement>(null)
    const [isOpen, setIsOpen] = useState(false)

    return (
        <>
            <button ref={targetButtonReference} type="button" className="badge badge-secondary">
                peek
            </button>
            <Popover isOpen={isOpen} target={targetButtonReference} onVisibilityChange={setIsOpen}>
                {/* eslint-disable-next-line react/forbid-dom-props */}
                <div style={{ maxWidth: '75vw', maxHeight: '75vh' }} className="p-3">
                    <Markdown dangerousInnerHTML={file.richHTML} />
                </div>
            </Popover>
        </>
    )
}
