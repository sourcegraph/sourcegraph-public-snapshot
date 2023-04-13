import React, { useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'

import { Client, ClientInit, createClient } from '@sourcegraph/cody-shared/src/chat/client'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { Chat, ChatUISubmitButtonProps, ChatUITextAreaProps } from '@sourcegraph/cody-ui/src/Chat'
import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'
import { Terms } from '@sourcegraph/cody-ui/src/Terms'
import { SubmitSvg } from '@sourcegraph/cody-ui/src/utils/icons'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    TextArea as WildcardTextArea,
    Alert,
    ErrorAlert,
    Form,
    Label,
    LoadingSpinner,
    Text,
} from '@sourcegraph/wildcard'

import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'
import { RepoEmbeddingExistsQueryResult, RepoEmbeddingExistsQueryVariables } from '../../../graphql-operations'
import { REPO_EMBEDDING_EXISTS_QUERY } from '../../../repo/repoRevisionSidebar/cody/backend'
import { eventLogger } from '../../../tracking/eventLogger'
import { RepositoryField } from '../../insights/components'

import codyStyles from '../../../repo/repoRevisionSidebar/cody/RepoRevisionSidebarCody.module.scss'
import styles from './CodyHomepage.module.scss'

interface CodeSearchPageProps {
    authenticatedUser: AuthenticatedUser | null
    telemetryService: TelemetryService
}

export const CodyHomepage: React.FunctionComponent<CodeSearchPageProps> = ({ authenticatedUser }) => {
    useEffect(() => {
        eventLogger.logPageView('CodyHome')
    }, [])

    const [codyEnabled] = useFeatureFlag('cody-experimental', true)

    const [repository, setRepository] = useState<string>('github.com/hashicorp/errwrap')

    return (
        <div className={classNames('d-flex flex-column align-items-center mx-auto p-3 container', styles.container)}>
            {codyEnabled ? (
                <Inner repository={repository} onRepositoryChange={setRepository} />
            ) : (
                <Alert variant="info" className="mt-5">
                    Cody is not enabled on this Sourcegraph instance.
                </Alert>
            )}
        </div>
    )
}

const Inner: React.FunctionComponent<{
    repository: string | undefined
    onRepositoryChange: (value: string) => void
}> = ({ repository, onRepositoryChange }) => (
    <div className="w-100 d-flex flex-1 flex-column align-items-center">
        <Form className="w-100">
            <Label htmlFor="repository">Repository</Label>
            <RepositoryField
                id="repository"
                value={repository ?? ''}
                onChange={onRepositoryChange}
                className={classNames('mx-auto', styles.repositoryInput)}
            />
        </Form>
        {repository && <CodyUI repository={repository} />}
    </div>
)

const CodyUI: React.FunctionComponent<{
    repository: string
}> = ({ repository }) => {
    const config = useMemo<ClientInit['config']>(
        () => ({
            serverEndpoint: window.location.origin,
            useContext: 'embeddings',
            codebase: repository,
        }),
        [repository]
    )
    const [messageInProgress, setMessageInProgress] = useState<ChatMessage | null>(null)
    const [transcript, setTranscript] = useState<ChatMessage[]>([])
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])

    const [client, setClient] = useState<Client | ErrorLike>()
    useEffect(() => {
        setMessageInProgress(null)
        setTranscript([])
        createClient({ config, accessToken: null, setMessageInProgress, setTranscript }).then(setClient, error => {
            eventLogger.log('web:codyHome:clientError', { repo: repository })
            setClient(error)
        })
    }, [config, repository])

    const onSubmit = useCallback(
        (text: string) => {
            if (client && !isErrorLike(client)) {
                eventLogger.log('web:codyHome:submit', { repo: repository, text })
                client.submitMessage(text)
            }
        },
        [client, repository]
    )

    const { data, loading, error } = useQuery<RepoEmbeddingExistsQueryResult, RepoEmbeddingExistsQueryVariables>(
        REPO_EMBEDDING_EXISTS_QUERY,
        { variables: { repoName: repository } }
    )

    if (!client || loading) {
        return <LoadingSpinner />
    }

    if (error) {
        return <ErrorAlert error={error} />
    }

    if (isErrorLike(client)) {
        return <ErrorAlert error={client} />
    }

    if (!data?.repository?.embeddingExists) {
        return <Text>Repo embeddings not generated.</Text>
    }

    return (
        <Chat
            messageInProgress={messageInProgress}
            transcript={transcript}
            formInput={formInput}
            setFormInput={setFormInput}
            inputHistory={inputHistory}
            setInputHistory={setInputHistory}
            onSubmit={onSubmit}
            textAreaComponent={TextArea}
            submitButtonComponent={SubmitButton}
            fileLinkComponent={FileLink}
            className={classNames(styles.chatContainer, 'border my-3 pt-2 flex-1')}
            afterTips={
                <details className="small mt-2">
                    <summary>Terms</summary>
                    <Terms />
                </details>
            }
            bubbleContentClassName={codyStyles.bubbleContent}
            humanBubbleContentClassName={codyStyles.humanBubbleContent}
            botBubbleContentClassName={codyStyles.botBubbleContent}
            bubbleFooterClassName="text-muted small"
            bubbleLoaderDotClassName={codyStyles.bubbleLoaderDot}
            inputRowClassName={codyStyles.inputRow}
            chatInputClassName={codyStyles.chatInput}
        />
    )
}

const TextArea: React.FunctionComponent<ChatUITextAreaProps> = ({
    className,
    rows,
    autoFocus,
    value,
    required,
    onInput,
    onKeyDown,
}) => (
    <WildcardTextArea
        className={className}
        rows={rows}
        value={value}
        autoFocus={autoFocus}
        required={required}
        onInput={onInput}
        onKeyDown={onKeyDown}
    />
)

const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = ({ className, disabled, onClick }) => (
    <button className={className} type="submit" disabled={disabled} onClick={onClick}>
        <SubmitSvg />
    </button>
)

const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>
