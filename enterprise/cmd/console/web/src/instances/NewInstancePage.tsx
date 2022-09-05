import { Container } from '@sourcegraph/wildcard'
import React from 'react'
import styles from './NewInstancePage.module.scss'
import { Link } from 'react-router-dom'
import { NewInstanceForm } from './NewInstanceForm'

export const NewInstancePage: React.FunctionComponent<{}> = () => {
    return (
        <div className={styles.page}>
            <Container className={styles.container}>
                <h2 className="mt-2">Create new workspace</h2>
                <p className="text-muted mb-4">
                    If your organization is already using Sourcegraph, you can <Link to="TODO">sign in</Link>.
                </p>
                <NewInstanceForm />
            </Container>
        </div>
    )
}
