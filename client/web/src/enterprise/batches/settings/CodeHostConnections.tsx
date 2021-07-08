import React from 'react'

import { Container, PageHeader } from '@sourcegraph/wildcard'

import { Scalars } from '../../../graphql-operations'

import { CodeHostConnectionNodes, GlobalCodeHostConnectionNodes } from './CodeHostConnectionNodes'

export interface CodeHostConnectionsProps {
    userID: Scalars['ID'] | null
    headerLine: JSX.Element
}

export const CodeHostConnections: React.FunctionComponent<CodeHostConnectionsProps> = ({ userID, headerLine }) => (
    <>
        <PageHeader headingElement="h2" path={[{ text: 'Batch Changes' }]} className="mb-3" />
        <Container>
            <h3>Code host tokens</h3>
            {headerLine}
            {userID ? <CodeHostConnectionNodes userID={userID} /> : <GlobalCodeHostConnectionNodes />}
            <p className="mb-0">
                Code host not present? Site admins can add a code host in{' '}
                <a href="https://docs.sourcegraph.com/admin/external_service" target="_blank" rel="noopener noreferrer">
                    the manage repositories settings
                </a>
                .
            </p>
        </Container>
    </>
)
