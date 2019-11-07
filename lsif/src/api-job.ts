import { Job } from 'bull'
import { ApiJobState } from './queue'

/**
 * The representation of a job as returned by the API.
 */
export interface ApiJob {
    id: string
    name: string
    args: object
    state: ApiJobState
    progress: number
    failedReason: string | null
    stacktrace: string[] | null
    timestamp: string
    processedOn: string | null
    finishedOn: string | null
}

/**
 * Convert a timestamp into an ISO string.
 *
 * @param timestamp The millisecond POSIX timestamp.
 */

const toDate = (timestamp: number): string => new Date(timestamp).toISOString()

/**
 * Attempt to convert a timestamp into an ISO string.
 *
 * @param timestamp The millisecond POSIX timestamp.
 */
const toMaybeDate = (timestamp: number | null): string | null => (timestamp ? toDate(timestamp) : null)

/**
 * Attempt to convert a string into an integer.
 *
 * @param value The int-y string.
 */
const toMaybeInt = (value: string | undefined): number | null => (value ? parseInt(value, 10) : null)

/**
 * Format a job to return from the API.
 *
 * @param job The job to format.
 * @param state The job's state.
 */
export const formatJob = (job: Job, state: ApiJobState): ApiJob => {
    const payload = job.toJSON()

    return {
        id: `${payload.id}`,
        name: payload.name,
        args: payload.data.args,
        state,
        progress: payload.progress,
        failedReason: payload.failedReason,
        stacktrace: payload.stacktrace,
        timestamp: toDate(payload.timestamp),
        processedOn: toMaybeDate(payload.processedOn),
        finishedOn: toMaybeDate(payload.finishedOn),
    }
}

/**
 * Format a job to return from the API.
 *
 * @param values A map of values composing the job.
 * @param state The job's state.
 */
export const formatJobFromMap = (values: Map<string, string>, state: ApiJobState): ApiJob => {
    const rawData = values.get('data')
    const rawStacktrace = values.get('stacktrace')

    return {
        id: values.get('id') || '',
        name: values.get('name') || '',
        args: rawData ? JSON.parse(rawData).args : {},
        state,
        progress: toMaybeInt(values.get('progress')) || 0,
        failedReason: values.get('failedReason') || null,
        stacktrace: rawStacktrace ? JSON.parse(rawStacktrace) : null,
        timestamp: toMaybeDate(toMaybeInt(values.get('timestamp'))) || '',
        processedOn: toMaybeDate(toMaybeInt(values.get('processedOn'))),
        finishedOn: toMaybeDate(toMaybeInt(values.get('finishedOn'))),
    }
}
