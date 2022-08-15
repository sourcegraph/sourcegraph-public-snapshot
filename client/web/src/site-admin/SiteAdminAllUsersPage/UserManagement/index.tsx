import React, { useEffect } from 'react'

// TODO: Fix typos, linting, types, self-refactoring

import { mdiAccount, mdiPlus, mdiDownload } from '@mdi/js'
import { RouteComponentProps } from 'react-router'

import { H1, Card, Text, Icon, Button, Link } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../tracking/eventLogger'

import { UsersList } from './components/UsersList'
import { UsersSummary } from './components/UsersSummary'

import styles from './index.module.scss'

export const UsersManagement: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    useEffect(() => {
        eventLogger.logPageView('UsersManagement')
    }, [])
    return (
        <>
            <div className="d-flex justify-content-between align-items-center mb-4 mt-2">
                <H1 className="d-flex align-items-center mb-0">
                    <Icon
                        svgPath={mdiAccount}
                        aria-label="user administration avatar icon"
                        size="md"
                        className={styles.linkColor}
                    />{' '}
                    User administration
                </H1>
                <div>
                    <Button
                        href="/site-admin/usage-statistics/archive"
                        download="true"
                        className="mr-4"
                        variant="secondary"
                        outline={true}
                        as="a"
                    >
                        <Icon svgPath={mdiDownload} aria-label="Download usage stats" className="mr-1" />
                        Download usage stats
                    </Button>
                    <Button to="/site-admin/users/new" variant="primary" as={Link}>
                        <Icon svgPath={mdiPlus} aria-label="create user" className="mr-1" />
                        Create User
                    </Button>
                </div>
            </div>
            <Card className="p-3">
                <UsersSummary />
                <UsersList />
            </Card>
            <Text className="font-italic text-center mt-2">
                All events are generated from entries in the event logs table and are updated every 24 hours..
            </Text>
        </>
    )
}
