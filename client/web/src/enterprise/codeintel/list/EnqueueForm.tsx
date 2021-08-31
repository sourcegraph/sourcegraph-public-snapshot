import React, { FunctionComponent, useCallback, useState } from 'react'
import { Subject } from 'rxjs'

import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Button } from '@sourcegraph/wildcard'

import { enqueueIndexJob as defaultEnqueueIndexJob } from './backend'

export interface EnqueueFormProps {
    repoId: string
    querySubject: Subject<string>
    enqueueIndexJob: typeof defaultEnqueueIndexJob
}

enum State {
    Idle,
    Queueing,
    Queued,
}

export const EnqueueForm: FunctionComponent<EnqueueFormProps> = ({ repoId, querySubject, enqueueIndexJob }) => {
    const [revlike, setRevlike] = useState('HEAD')
    const [state, setState] = useState(() => State.Idle)
    const [queueResult, setQueueResult] = useState<number>()
    const [enqueueError, setEnqueueError] = useState<Error>()

    const enqueue = useCallback(async () => {
        setState(State.Queueing)
        setEnqueueError(undefined)
        setQueueResult(undefined)

        try {
            const indexes = await enqueueIndexJob(repoId, revlike).toPromise()
            setQueueResult(indexes.length)
            if (indexes.length > 0) {
                querySubject.next(indexes[0].inputCommit)
            }
        } catch (error) {
            setEnqueueError(error)
            setQueueResult(undefined)
        } finally {
            setState(State.Queued)
        }
    }, [repoId, revlike, querySubject, enqueueIndexJob])

    return (
        <>
            {enqueueError && <ErrorAlert prefix="Error enqueueing index job" error={enqueueError} />}

            <div className="form-inline">
                <label htmlFor="revlike">Git revlike</label>

                <input
                    type="text"
                    id="revlike"
                    className="form-control ml-2"
                    value={revlike}
                    onChange={event => setRevlike(event.target.value)}
                />

                <Button
                    type="button"
                    title="Enqueue thing"
                    disabled={state === State.Queueing}
                    className="ml-2"
                    variant="primary"
                    onClick={enqueue}
                >
                    Enqueue
                </Button>
            </div>

            {state === State.Queued && queueResult !== undefined && (
                <div className="alert alert-success mt-3 mb-0">{queueResult} index jobs enqueued.</div>
            )}
        </>
    )
}
