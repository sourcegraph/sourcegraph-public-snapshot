import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React, { useRef, useState } from 'react'
import { Link } from 'react-router-dom'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { CatalogComponentDocumentationFields, CatalogEntityDetailFields } from '../../../../../graphql-operations'
import { formatPersonName, PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { Popover } from '../../../../insights/components/popover/Popover'
import { CatalogEntityIcon } from '../../../components/CatalogEntityIcon'
import { EntityOwner } from '../../../components/entity-owner/EntityOwner'

import { ComponentSourceDefinitions } from './ComponentSourceDefinitions'
import { EntityCatalogExplorer } from './EntityCatalogExplorer'
import { OverviewStatusContexts } from './OverviewStatusContexts'

interface Props {
    entity: CatalogEntityDetailFields
    className?: string
}

export const EntityOverviewTab: React.FunctionComponent<Props> = ({ entity, className }) => (
    <div className="flex-1 align-self-stretch row no-gutters">
        <div className="col-md-4 col-lg-3 col-xl-2 border-right p-3">
            {entity.name && (
                <h2 className="d-flex align-items-center mb-1">
                    <CatalogEntityIcon entity={entity} className="icon-inline mr-2" /> {entity.name}
                </h2>
            )}
            <div className="text-muted small mb-2">
                {entity.__typename === 'CatalogComponent' && `${entity.kind[0]}${entity.kind.slice(1).toLowerCase()}`}
            </div>
            {entity.description && <p className="mb-3">{entity.description}</p>}
            <div>
                <Link
                    to={`/search?q=context:c/${entity.name}`}
                    className="d-inline-flex align-items-center btn btn-outline-secondary mb-3"
                >
                    <SearchIcon className="icon-inline mr-1" /> Search...
                </Link>
                {entity.readme && (
                    <div className="d-flex align-items-start">
                        <Link to={entity.readme.url} className="d-flex align-items-center text-body mb-3 mr-2">
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
                <dl>
                    <dt>Owner</dt>
                    <dd>
                        <EntityOwner owner={entity.owner} className="d-block" />
                        <ul className="list-unstyled d-flex flex-wrap">
                            {entity.owner?.__typename === 'Group' &&
                                entity.owner.members.map(member => (
                                    <li key={member.email} className="mr-1 mb-1">
                                        <LinkOrSpan to={member.user?.url} title={formatPersonName(member)}>
                                            <UserAvatar user={member} size={18} />
                                        </LinkOrSpan>
                                    </li>
                                ))}
                        </ul>
                    </dd>
                    <dt>Lifecycle</dt>
                    <dd>{entity.lifecycle.toLowerCase()}</dd>
                </dl>
            </div>
        </div>
        <div className="col-md-8 col-lg-9 col-xl-10 p-3">
            <div className="card mb-3">
                <ComponentSourceDefinitions catalogComponent={entity} listGroupClassName="list-group-flush" />
                {entity.commits?.nodes[0] && <LastCommit commit={entity.commits?.nodes[0]} className="card-footer" />}
            </div>
            <OverviewStatusContexts entity={entity} itemClassName="mb-3" />
            <EntityCatalogExplorer entity={entity.id} className="mb-3" />
        </div>
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
