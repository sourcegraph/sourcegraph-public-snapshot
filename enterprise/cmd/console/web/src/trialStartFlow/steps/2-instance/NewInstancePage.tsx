import { Container } from '@sourcegraph/wildcard'
import React from 'react'
import styles from './NewInstancePage.module.scss'
import { Link } from 'react-router-dom'
import { NewInstanceForm } from '../../../instances/NewInstanceForm'
import { OrganizationExistingInstances } from '../../../instances/OrganizationExistingInstances'
import { TrialStartFlowContainer } from '../../TrialStartFlowContainer'

export const NewInstancePage: React.FunctionComponent<{}> = () => {
    const hasOrganizationExistingInstances = true
    return (
        <TrialStartFlowContainer>
            <h3 className="mt-4">Select instance</h3>
            <p className="text-muted">
                Your email domain <strong>sourcegraph.com</strong> already has 2 instances that you may be able to sign
                into.
            </p>
            <OrganizationExistingInstances />
            <section className="mt-5">
                <h3>Or create a new instance</h3>
                <p className="text-muted">After it's created, you can add repositories and invite other people.</p>
                <NewInstanceForm />
                <p className="text-muted mt-3 mb-0">
                    Want to <Link to="#">create a self-managed Sourcegraph instance</Link> on your own infrastructure?
                    {/* TODO(sqs): or have a way to select cloud vs. self-hosted above */}
                </p>
            </section>
        </TrialStartFlowContainer>
    )
}
