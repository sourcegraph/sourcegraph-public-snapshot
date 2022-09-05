import { Container } from '@sourcegraph/wildcard'
import React from 'react'
import styles from './NewInstancePage.module.scss'
import { Link } from 'react-router-dom'
import { NewInstanceForm } from './NewInstanceForm'
import { OrganizationExistingInstances } from './OrganizationExistingInstances'

export const NewInstancePage: React.FunctionComponent<{}> = () => {
    const hasOrganizationExistingInstances = true
    return (
        <div className={styles.page}>
            <Container className={styles.container}>
                <h2 className="mt-2">Select instance</h2>
                <OrganizationExistingInstances />
                <h3>Or create a new instance</h3>
                <NewInstanceForm />
            </Container>
        </div>
    )
}
