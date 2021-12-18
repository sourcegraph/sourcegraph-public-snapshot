import classNames from 'classnames'
import AccountIcon from 'mdi-react/AccountIcon'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { GroupDetailFields } from '../../../../graphql-operations'
import { formatPersonName } from '../../../../person/PersonLink'
import { UserAvatar } from '../../../../user/UserAvatar'
import { ComponentIcon } from '../../components/ComponentIcon'
import { CatalogGroupIcon } from '../../components/CatalogGroupIcon'
import { ComponentStateIndicator } from '../overview/components/entity-state-indicator/EntityStateIndicator'

import { GroupCatalogExplorer } from './GroupCatalogExplorer'
import { GroupDetailContentCardProps } from './GroupDetailContent'
import { GroupLink } from './GroupLink'
import styles from './GroupOverviewTab.module.scss'

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
    <div className={classNames('flex-1 row no-gutters', className)}>
        <div className="col-md-4 col-lg-3 col-xl-2 border-right p-3">
            <h2 className="d-flex align-items-center mb-1">
                <CatalogGroupIcon className="icon-inline mr-2" />
                {group.title || group.name}
            </h2>
            <div className="text-muted small mb-2">Group</div>
            {group.description && <p className="mb-3">{group.description}</p>}
            <Link
                to={`/search?q=context:g/${group.name}`}
                className="d-inline-flex align-items-center btn btn-outline-secondary mb-3"
            >
                <SearchIcon className="icon-inline mr-1" /> Search code...
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
            {group.members && group.members.length > 0 && (
                <>
                    <h4 className="font-weight-bold">
                        <Link to={`${group.url}/members`} className="text-body">
                            {group.members.length} {pluralize('member', group.members.length)}
                        </Link>
                    </h4>
                    <ul className="list-unstyled d-flex flex-wrap">
                        {group.members.map(member => (
                            <li key={member.email} className="mr-1 mb-1">
                                <LinkOrSpan to={member.user?.url} title={formatPersonName(member)}>
                                    <UserAvatar user={member} size={28} />
                                </LinkOrSpan>
                            </li>
                        ))}
                    </ul>
                </>
            )}
        </div>
        <div className="col-md-8 col-lg-9 col-xl-10 p-3">
            {group.childGroups && group.childGroups.length > 0 && (
                <div className="mb-3">
                    <h4 className="font-weight-bold">
                        {group.childGroups.length} {pluralize('subgroup', group.childGroups.length)}
                    </h4>
                    <ul className={styles.boxGrid}>
                        {group.childGroups.map(childGroup => (
                            <li
                                key={childGroup.id}
                                className={classNames(
                                    'position-relative border rounded d-flex flex-column',
                                    styles.boxGridItem
                                )}
                            >
                                <GroupLink group={childGroup} className="stretched-link" />
                                {childGroup.description && (
                                    <p className={classNames('my-1 text-muted small', styles.boxGridItemBody)}>
                                        {childGroup.description}
                                    </p>
                                )}
                                <div className="flex-1" />
                                <div className="text-muted small">
                                    <AccountIcon className="icon-inline" /> {childGroup.members.length}{' '}
                                    {pluralize('member', childGroup.members.length)}
                                </div>
                            </li>
                        ))}
                    </ul>
                </div>
            )}

            <GroupCatalogExplorer group={group.id} className="mb-3" />

            {group.ownedEntities && group.ownedEntities.length > 0 && false && (
                <div className="mb-3">
                    <h4 className="font-weight-bold">
                        {group.ownedEntities.length} {pluralize('component', group.ownedEntities.length)}
                    </h4>
                    <ul className={styles.boxGrid}>
                        {group.ownedEntities.map(entity => (
                            <li
                                key={entity.id}
                                className={classNames(
                                    'position-relative border rounded d-flex flex-column',
                                    styles.boxGridItem
                                )}
                            >
                                <div className="mb-0 d-flex align-items-center">
                                    <Link to={entity.url} className="d-inline-flex align-items-center stretched-link">
                                        <ComponentIcon entity={entity} className="icon-inline text-muted mr-1" />

                                        {entity.name}
                                    </Link>
                                    {entity.__typename === 'Component' && (
                                        <ComponentStateIndicator entity={entity} className="ml-1" />
                                    )}
                                </div>
                                {entity.description && (
                                    <p className={classNames('my-1 text-muted small', styles.boxGridItemBody)}>
                                        {entity.description}
                                    </p>
                                )}
                            </li>
                        ))}
                    </ul>
                </div>
            )}

            {group.ownedEntities && group.ownedEntities.length > 0 && false && (
                <div className="card mb-3">
                    <header className={classNames(headerClassName)}>
                        <h4 className={classNames('mb-0 mr-2', titleClassName)}>Components</h4>
                    </header>
                    <ul className="list-group list-group-flush">
                        {group.ownedEntities.map(entity => (
                            <li key={entity.id} className="list-group-item d-flex align-items-center position-relative">
                                <Link to={entity.url} className="d-flex align-items-center mr-1">
                                    <ComponentIcon entity={entity} className="icon-inline text-muted mr-1" />
                                    {entity.name}
                                </Link>
                                {entity.__typename === 'Component' && (
                                    <ComponentStateIndicator entity={entity} />
                                )}
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    </div>
)
