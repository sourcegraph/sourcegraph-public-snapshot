import { useEffect, type FunctionComponent } from 'react'

import classNames from 'classnames'
import { useLocation, useNavigate, useParams } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Alert, Button, Container, ErrorAlert, H3, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import type { WorkflowFields, WorkflowResult, WorkflowVariables } from '../graphql-operations'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'

import { WORKFLOW_UPDATED_LOCATION_STATE_KEY } from './EditPage'
import { workflowQuery } from './graphql'
import { WorkflowPage } from './Page'
import { WorkflowNameWithOwner } from './WorkflowNameWithOwner'

import styles from './DetailPage.module.scss'

/**
 * Page to show a workflow.
 */
export const DetailPage: FunctionComponent<TelemetryV2Props> = ({ telemetryRecorder }) => {
    const { id } = useParams<{ id: string }>()

    const { data, loading, error } = useQuery<WorkflowResult, WorkflowVariables>(workflowQuery, {
        variables: { id: id! },
    })
    const workflow = data?.node?.__typename === 'Workflow' ? data.node : null

    // Flash after updating.
    const location = useLocation()
    const navigate = useNavigate()
    const justUpdated = !!location.state?.[WORKFLOW_UPDATED_LOCATION_STATE_KEY]
    useEffect(() => {
        if (justUpdated) {
            setTimeout(() => navigate({}, { state: {} }), 1000)
        }
    }, [justUpdated, navigate])
    const flash = justUpdated ? 'Saved!' : null

    return (
        <WorkflowPage
            title={workflow ? `${workflow.description} - workflow` : 'Workflow'}
            actions={
                workflow?.viewerCanAdminister && (
                    <Button to={`${workflow.url}/edit`} variant="secondary" as={Link}>
                        Edit
                    </Button>
                )
            }
            breadcrumbsNamespace={workflow?.owner}
            breadcrumbs={workflow ? <PageHeader.Breadcrumb>{workflow.description}</PageHeader.Breadcrumb> : null}
        >
            {loading ? (
                <LoadingSpinner />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : !workflow ? (
                <Alert variant="danger" as="p">
                    Workflow not found.
                </Alert>
            ) : (
                <>
                    <Detail workflow={workflow} telemetryRecorder={telemetryRecorder} />
                    {flash && !loading && (
                        <Alert variant="success" className="my-3">
                            {flash}
                        </Alert>
                    )}
                </>
            )}
        </WorkflowPage>
    )
}

const Detail: FunctionComponent<TelemetryV2Props & { workflow: WorkflowFields }> = ({
    workflow,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('workflows.detail', 'view', {
            metadata: namespaceTelemetryMetadata(workflow.owner),
        })
    }, [telemetryRecorder, workflow.owner])

    return (
        <Container className={classNames(styles.container)}>
            <div className="d-flex flex-column flex-gap-2 align-items-center">
                <H3>
                    <WorkflowNameWithOwner workflow={workflow} />{' '}
                </H3>
                TODO!(sqs): show description and template
            </div>
        </Container>
    )
}
