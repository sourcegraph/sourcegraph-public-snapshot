import * as H from 'history'
import AccountIcon from 'mdi-react/AccountIcon'
import AddIcon from 'mdi-react/AddIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject } from 'rxjs'

import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { Button } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../components/alerts'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { OrganizationFields } from '../graphql-operations'
import { orgURL } from '../org'

import { deleteOrganization, fetchAllOrganizations } from './backend'

interface OrgNodeProps {
    /**
     * The org to display in this list item.
     */
    node: OrganizationFields

    /**
     * Called when the org is updated by an action in this list item.
     */
    onDidUpdate?: () => void
    history: H.History
}

const OrgNode: React.FunctionComponent<OrgNodeProps> = ({ node, onDidUpdate }) => {
    const [loading, setLoading] = useState<boolean | Error>(false)

    const deleteOrg = useCallback(() => {
        if (!window.confirm(`Delete the organization ${node.name}?`)) {
            return
        }

        setLoading(true)

        deleteOrganization(node.id).then(
            () => {
                setLoading(false)
                if (onDidUpdate) {
                    onDidUpdate()
                }
            },
            error => setLoading(asError(error))
        )
    }, [node.id, node.name, onDidUpdate])

    return (
        <li className="list-group-item py-2">
            <div className="d-flex align-items-center justify-content-between">
                <div>
                    <Link to={orgURL(node.name)}>
                        <strong>{node.name}</strong>
                    </Link>
                    <br />
                    <span className="text-muted">{node.displayName}</span>
                </div>
                <div>
                    <Button
                        to={`${orgURL(node.name)}/settings`}
                        data-tooltip="Organization settings"
                        variant="secondary"
                        size="sm"
                        as={Link}
                    >
                        <SettingsIcon className="icon-inline" /> Settings
                    </Button>{' '}
                    <Button
                        to={`${orgURL(node.name)}/settings/members`}
                        data-tooltip="Organization members"
                        variant="secondary"
                        size="sm"
                        as={Link}
                    >
                        <AccountIcon className="icon-inline" />{' '}
                        {node.members && (
                            <>
                                {node.members.totalCount} {pluralize('member', node.members.totalCount)}
                            </>
                        )}
                    </Button>{' '}
                    <Button
                        onClick={deleteOrg}
                        disabled={loading === true}
                        data-tooltip="Delete organization"
                        variant="danger"
                        size="sm"
                    >
                        <DeleteIcon className="icon-inline" />
                    </Button>
                </div>
            </div>
            {isErrorLike(loading) && <ErrorAlert className="mt-2" error={loading.message} />}
        </li>
    )
}

interface Props extends RouteComponentProps<{}>, TelemetryProps {}

/**
 * A page displaying the orgs on this site.
 */
export const SiteAdminOrgsPage: React.FunctionComponent<Props> = ({ telemetryService, history, location }) => {
    const orgUpdates = useMemo(() => new Subject<void>(), [])
    const onDidUpdateOrg = useCallback((): void => orgUpdates.next(), [orgUpdates])

    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminOrgs')
    }, [telemetryService])

    return (
        <div className="site-admin-orgs-page">
            <PageTitle title="Organizations - Admin" />
            <div className="d-flex justify-content-between align-items-center mb-3">
                <h2 className="mb-0">Organizations</h2>
                <Button to="/organizations/new" className="test-create-org-button" variant="primary" as={Link}>
                    <AddIcon className="icon-inline" /> Create organization
                </Button>
            </div>
            <p>
                An organization is a set of users with associated configuration. See{' '}
                <Link to="/help/admin/organizations">Sourcegraph documentation</Link> for information about configuring
                organizations.
            </p>
            {window.context.sourcegraphDotComMode ? (
                <>
                    <div className="alert alert-info">
                        Only organization members can view & modify organization settings.
                    </div>
                    <h3>Enable early access</h3>
                    <div className="d-flex justify-content-between align-items-center mb-3">
                        <p>Enable early access for organization code host connections and repositories on Cloud.</p>
                        <Button to="./organizations/early-access-orgs-code" variant="primary" outline={true} as={Link}>
                            Enable early access
                        </Button>
                    </div>
                </>
            ) : (
                <FilteredConnection<OrganizationFields, Omit<OrgNodeProps, 'node'>>
                    className="list-group list-group-flush mt-3"
                    noun="organization"
                    pluralNoun="organizations"
                    queryConnection={fetchAllOrganizations}
                    nodeComponent={OrgNode}
                    nodeComponentProps={{
                        onDidUpdate: onDidUpdateOrg,
                        history,
                    }}
                    updates={orgUpdates}
                    history={history}
                    location={location}
                />
            )}
        </div>
    )
}
