import React, { useEffect } from 'react'

import classNames from 'classnames'

import { Maybe } from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ButtonLink, Container, Link, PageHeader, Typography } from '@sourcegraph/wildcard'

import { refreshAuthenticatedUser } from '../../../auth'
import { eventLogger } from '../../../tracking/eventLogger'

import styles from './organizationList.module.scss'
export interface OrganizationsListProps extends ThemeProps {
    authenticatedUser: Pick<
        AuthenticatedUser,
        'id' | 'username' | 'avatarURL' | 'settingsURL' | 'organizations' | 'siteAdmin' | 'session' | 'displayName'
    >
}

interface IOrgItem {
    id: string
    name: string
    displayName: Maybe<string>
    url: string
    settingsURL: Maybe<string>
}

interface OrgItemProps {
    org: IOrgItem
}

const OrgItem: React.FunctionComponent<React.PropsWithChildren<OrgItemProps>> = ({ org }) => (
    <li data-test-username={org.id}>
        <div className={classNames('d-flex align-items-center justify-content-start flex-1', styles.orgDetails)}>
            <div className={styles.avatarContainer}>
                <div className={styles.avatar}>
                    <span>{(org.displayName || org.name).slice(0, 2).toUpperCase()}</span>
                </div>
            </div>
            <div className="d-flex flex-column">
                <Link to={`${org.url}/getstarted`} className={styles.orgLink}>
                    {org.displayName || org.name}
                </Link>
                {org.displayName && <span className={classNames('text-muted', styles.displayName)}>{org.name}</span>}
            </div>
        </div>

        <div className={styles.userRole}>
            <span className="text-muted">Admin</span>
        </div>
        <div>
            <ButtonLink
                className={styles.orgSettings}
                variant="secondary"
                to={org.settingsURL as string}
                size="sm"
                onClick={() =>
                    eventLogger.log(
                        'UserOrganizationSettingsClicked',
                        { organizationId: org.id },
                        { organizationId: org.id }
                    )
                }
            >
                Settings
            </ButtonLink>
        </div>
    </li>
)

const refreshOrganizationList = (): void => {
    refreshAuthenticatedUser()
        .toPromise()
        .then(() => {
            eventLogger.logViewEvent('OrganizationsList')
        })
        .catch(() => eventLogger.logViewEvent('ErrorOrgListLoading'))
}

export const OrganizationsListPage: React.FunctionComponent<React.PropsWithChildren<OrganizationsListProps>> = ({
    authenticatedUser,
}) => {
    useEffect(() => {
        refreshOrganizationList()
    }, [])

    const orgs = authenticatedUser.organizations.nodes
    const hasOrgs = orgs.length > 0

    useEffect(() => {
        eventLogger.logPageView('YourOrganizations', { userId: authenticatedUser.id })
    }, [authenticatedUser.id])

    return (
        <div className="org-list-page">
            <div className="d-flex flex-0 justify-content-end align-items-center mb-3 flex-wrap">
                <PageHeader path={[{ text: 'Organizations' }]} headingElement="h2" className="flex-1" />

                <ButtonLink
                    variant="secondary"
                    to="/organizations/joinopenbeta"
                    onClick={() => eventLogger.log('CreateOrganizationButtonClicked')}
                >
                    Create organization
                </ButtonLink>
            </div>
            {hasOrgs && (
                <Container className={classNames('mb-4', styles.organisationsList)}>
                    <ul>
                        {orgs.map(org => (
                            <OrgItem org={org} key={org.id} />
                        ))}
                    </ul>
                </Container>
            )}
            {!hasOrgs && (
                <Container className={styles.noOrgContainer}>
                    <div className="d-flex flex-0 flex-column justify-content-center align-items-center">
                        <Typography.H3 className="mb-1">Start searching with your team on Sourcegraph</Typography.H3>
                        <div>Level up your team with powerful code search across your organization’s code.</div>
                        <ButtonLink
                            variant="primary"
                            to="/organizations/joinopenbeta"
                            className="mt-3"
                            onClick={() => eventLogger.log('CreateFirstOrganizationButtonClicked')}
                        >
                            Create your first organization
                        </ButtonLink>
                    </div>
                </Container>
            )}
        </div>
    )
}
