import { useEffect, type FunctionComponent } from 'react'

import { useLocation, useNavigate, useParams } from 'react-router-dom'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { useQuery } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { Alert, Button, Container, ErrorAlert, H5, Link, LoadingSpinner, PageHeader, Text } from '@sourcegraph/wildcard'

import type { PromptFields, PromptResult, PromptVariables } from '../graphql-operations'
import { LibraryItemStatusBadge, LibraryItemVisibilityBadge } from '../library/itemBadges'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'

import { PROMPT_UPDATED_LOCATION_STATE_KEY } from './EditPage'
import { promptQuery } from './graphql'
import { PromptPage } from './Page'
import { urlToEditPrompt } from './util'

import styles from './DetailPage.module.scss'

/**
 * Page to show a prompt.
 */
export const DetailPage: FunctionComponent<TelemetryV2Props> = ({ telemetryRecorder }) => {
    const { id } = useParams<{ id: string }>()

    const { data, loading, error } = useQuery<PromptResult, PromptVariables>(promptQuery, {
        variables: { id: id! },
    })
    const prompt = data?.node?.__typename === 'Prompt' ? data.node : null

    // Flash after updating.
    const location = useLocation()
    const navigate = useNavigate()
    const justUpdated = !!location.state?.[PROMPT_UPDATED_LOCATION_STATE_KEY]
    useEffect(() => {
        if (justUpdated) {
            setTimeout(() => navigate({}, { state: {} }), 1000)
        }
    }, [justUpdated, navigate])
    const flash = justUpdated ? 'Saved!' : null

    return (
        <PromptPage
            title={prompt ? `Prompt ${prompt.nameWithOwner}` : 'Prompt'}
            actions={
                prompt?.viewerCanAdminister && (
                    <Button to={urlToEditPrompt(prompt)} variant="secondary" as={Link}>
                        Edit
                    </Button>
                )
            }
            breadcrumbsNamespace={prompt?.owner}
            breadcrumbs={prompt ? <PageHeader.Breadcrumb>{prompt.name}</PageHeader.Breadcrumb> : null}
        >
            {loading ? (
                <LoadingSpinner />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : !prompt ? (
                <Alert variant="danger" as="p">
                    Prompt not found.
                </Alert>
            ) : (
                <>
                    <Detail prompt={prompt} telemetryRecorder={telemetryRecorder} />
                    {flash && !loading && (
                        <Alert variant="success" className="my-3">
                            {flash}
                        </Alert>
                    )}
                </>
            )}
        </PromptPage>
    )
}

const Detail: FunctionComponent<TelemetryV2Props & { prompt: PromptFields }> = ({ prompt, telemetryRecorder }) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('prompts.detail', 'view', {
            metadata: namespaceTelemetryMetadata(prompt.owner),
        })
    }, [telemetryRecorder, prompt.owner])

    return (
        <>
            <Text>
                {prompt.description}
                {prompt.description ? ' — ' : ''}
                <LibraryItemVisibilityBadge item={prompt} />
                <LibraryItemStatusBadge item={prompt} className="ml-1" />{' '}
                <small>
                    Last updated <Timestamp date={prompt.updatedAt} noAbout={true} />
                    {prompt.updatedBy && (
                        <>
                            {' '}
                            by{' '}
                            <Link to={prompt.updatedBy.url}>
                                <strong>{prompt.updatedBy.username}</strong>
                            </Link>
                        </>
                    )}
                </small>
            </Text>
            <H5 className="mt-4 mb-2">Prompt template</H5>
            <Container>
                <Text className={styles.definitionText}>
                    {prompt.definition.text.trim() === '' ? '(empty)' : prompt.definition.text}
                </Text>
            </Container>
        </>
    )
}
