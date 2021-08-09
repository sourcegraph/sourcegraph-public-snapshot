import * as H from 'history'
import React, { FunctionComponent, useEffect, useCallback, useState } from 'react'
import { RouteComponentProps } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { enqueueIndexJob } from './backend'

export interface CodeIntelIndexScheduleConfigurationPageProps
    extends RouteComponentProps<{}>,
        ThemeProps,
        TelemetryProps {
    repo: { id: string }
    history: H.History
}

enum State {
    Idle,
    Queueing,
    Queued,
}

export const CodeIntelIndexScheduleConfigurationPage: FunctionComponent<CodeIntelIndexScheduleConfigurationPageProps> = ({
    repo,
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexScheduleConfigurationPage'), [telemetryService])

    const [enqueueError, setEnqueueError] = useState<Error>()
    const [state, setState] = useState(() => State.Idle)
    const [revlike, setRevlike] = useState('HEAD')

    const onClick = useCallback(async () => {
        setState(State.Queueing)
        setEnqueueError(undefined)

        try {
            await enqueueIndexJob(repo.id, revlike).toPromise()
        } catch (error) {
            setEnqueueError(error)
        } finally {
            setState(State.Queued)
        }
    }, [repo, revlike])

    return (
        <div className="code-intel-index-configuration">
            <PageTitle title="Auto-indexing schedule configuration" />

            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Auto-indexing schedule configuration</>,
                    },
                ]}
                description="TODO"
                className="mb-3"
            />

            <Container>
                <div>
                    {enqueueError && <ErrorAlert prefix="Error enqueueing index job" error={enqueueError} />}

                    <input type="text" value={revlike} onChange={event => setRevlike(event.target.value)} />

                    <button
                        type="button"
                        title="Enqueue thing"
                        disabled={state === State.Queueing}
                        className="btn btn-sm btn-secondary"
                        onClick={onClick}
                    >
                        Enqueue
                    </button>

                    {state === State.Queued && <div className="text-success">Index jobs enqueued</div>}
                </div>
            </Container>
        </div>
    )
}
