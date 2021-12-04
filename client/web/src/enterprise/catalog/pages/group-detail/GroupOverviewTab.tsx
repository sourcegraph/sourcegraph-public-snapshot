import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { GroupDetailFields } from '../../../../graphql-operations'
import { PersonLink } from '../../../../person/PersonLink'
import { CatalogEntityIcon } from '../../components/CatalogEntityIcon'
import { CatalogEntityStateIndicator } from '../overview/components/entity-state-indicator/EntityStateIndicator'

import { GroupDetailContentCardProps } from './GroupDetailContent'
import { GroupLink } from './GroupLink'

interface Props extends GroupDetailContentCardProps {
    group: GroupDetailFields
    className?: string
}

export const GroupOverviewTab: React.FunctionComponent<Props> = ({
    group,
    headerClassName,
    titleClassName,
    bodyClassName,
    className,
}) => (
    <div className={classNames('d-flex flex-column', className)}>
        <div className="row">
            <div className="col-md-3">
                {group.title && <h2>{group.title}</h2>}
                {group.description && <p className="mb-3">{group.description}</p>}
                {(group.parentGroup || (group.childGroups && group.childGroups.length > 0)) && (
                    <>
                        <dl className="mb-2">
                            {group.parentGroup && (
                                <>
                                    <dt className={classNames('mb-0 mr-2', titleClassName)}>Parent group</dt>
                                    <dd className={classNames('mb-0 d-flex align-items-center position-relative')}>
                                        <GroupLink group={group.parentGroup} className="stretched-link" />
                                    </dd>
                                </>
                            )}
                            {group.childGroups && group.childGroups.length > 0 && (
                                <>
                                    <dt className={classNames('mb-0 mr-2', titleClassName)}>Subgroups</dt>
                                    <dd className="mb-0">
                                        <ul className="list-unstyled mb-0">
                                            {group.childGroups.map(childGroup => (
                                                <li
                                                    key={childGroup.id}
                                                    className="d-flex align-items-center position-relative"
                                                >
                                                    <GroupLink group={childGroup} className="stretched-link" />
                                                </li>
                                            ))}
                                        </ul>
                                    </dd>
                                </>
                            )}
                        </dl>
                        <hr className="mb-3" />
                    </>
                )}

                <div>
                    <Link
                        to={`/search?q=context:g/${group.name}`}
                        className="d-inline-flex align-items-center btn btn-outline-secondary mb-3"
                    >
                        <SearchIcon className="icon-inline mr-1" /> Search group's code...
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body mb-3 mr-2">
                        <FileDocumentIcon className="icon-inline mr-2" />
                        Handbook page
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <AlertCircleOutlineIcon className="icon-inline mr-2" />
                        Issues
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body mb-3">
                        <SlackIcon className="icon-inline mr-2" />
                        #extensibility-chat
                    </Link>
                    <hr className="my-3" />
                </div>
            </div>
            <div className="col-md-9">
                {group.ownedEntities && group.ownedEntities.length > 0 && (
                    <div className="card mb-3">
                        <header className={classNames(headerClassName)}>
                            <h4 className={classNames('mb-0 mr-2', titleClassName)}>Components</h4>
                        </header>
                        <ul className="list-group list-group-flush">
                            {group.ownedEntities.map(entity => (
                                <li
                                    key={entity.id}
                                    className="list-group-item d-flex align-items-center position-relative"
                                >
                                    <Link to={entity.url} className="d-flex align-items-center mr-1">
                                        <CatalogEntityIcon entity={entity} className="icon-inline text-muted mr-1" />

                                        {entity.name}
                                    </Link>
                                    {entity.__typename === 'CatalogComponent' && (
                                        <CatalogEntityStateIndicator entity={entity} />
                                    )}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
                {group.members && group.members.length > 0 && (
                    <div className="card mb-3">
                        <header className={classNames(headerClassName)}>
                            <h4 className={classNames('mb-0 mr-2', titleClassName)}>Members</h4>
                        </header>
                        <ul className="list-group list-group-flush">
                            {group.members.map(member => (
                                <li
                                    key={member.email}
                                    className="list-group-item d-flex align-items-center position-relative"
                                >
                                    <PersonLink person={member} />
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
            </div>
        </div>
    </div>
)
