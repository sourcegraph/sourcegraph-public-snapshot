import React, { useEffect, FC, useCallback, useState } from 'react'

import { RouteComponentProps } from 'react-router'
import { mdiPlus, mdiChevronUp, mdiChevronDown } from '@mdi/js'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader, Button, Icon, Container, Text } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'

import styles from './SiteAdminRolesPage.module.scss'

export interface SiteAdminRolesPageProps extends RouteComponentProps, TelemetryProps {}

type Role = {
    id: number,
    name: string,
    system: boolean
}

export const SiteAdminRolesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminRolesPageProps>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminRoles')
    }, [telemetryService])

    const sampleRoles: Role[] = [
        {
            id: 1,
            name: 'TEST-ROLE-1',
            system: true
        },
        {
            id: 2,
            name: 'TEST-ROLE-2',
            system: true
        },
        {
            id: 3,
            name: 'TEST-ROLE-3',
            system: true
        },
        {
            id: 4,
            name: 'TEST-ROLE-4',
            system: true
        },
        {
            id: 5,
            name: 'TEST-ROLE-5',
            system: true
        },
    ]

    console.log('inside roles page')
    return (
        <div className="site-admin-roles-page">
            <PageTitle title="Roles - Admin" />
            <PageHeader
                path={[{ text: 'Roles' }]}
                headingElement="h2"
                description={
                    <>
                        A role is a set of permissions assigned to a user.
                    </>
                }
                className="mb-3"
                actions={
                    <Button variant="primary">
                        <Icon aria-hidden={true} svgPath={mdiPlus} /> Add Role
                    </Button>
                }
            />

            <Container>
                <ul className={styles.rolesList}>
                    {sampleRoles.map(role => <RoleNode node={role} key={role.id} />)}
                </ul>
            </Container>
        </div>
    )
}

const RoleNode: FC<{
    node: Role
}> = ({ node }) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    return (
        <li className={styles.roleNode}>
             <Button
                variant="icon"
                // className="d-none d-sm-block"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
            </Button>
            <Text className="font-weight-bold m-0">{node.name}</Text>
            {isExpanded ? (
                <div className={styles.roleNodePermissions}>
                    <Text>Grabbing Permissions</Text>
                    <Text>Grabbing Permissions</Text>
                    <Text>Grabbing Permissions</Text>
                    <Text>Grabbing Permissions</Text>
                    <Text>Grabbing Permissions</Text>
                    <Text>Grabbing Permissions</Text>
                </div>
            ) : <span />}
        </li>
    )
}


