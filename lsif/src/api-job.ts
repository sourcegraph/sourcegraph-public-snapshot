import { Job } from 'bull'

/**
 * The representation of a job as returned by the API.
 */
export interface ApiJob {
    id: string
    name: string
    args: object
    status: string
    progress: number
    timestamp: string
    finishedOn: string | null
    processedOn: string | null
    failedReason: string | null
    stacktrace: string[] | null
}

const toDate = (v: number): string => new Date(v).toISOString()
const toMaybeDate = (v: number | null): string | null => (v ? toDate(v) : null)
const toMaybeInt = (v: string | undefined) => (v ? parseInt(v, 10) : null)

/**
 * Format a job to return from the API.
 *
 * @param job The job to format.
 * @param status The job's status.
 */
export const formatJob = (job: Job, status: string): ApiJob => {
    const payload = job.toJSON()

    return {
        id: `${payload.id}`,
        name: payload.name,
        args: payload.data.args,
        status,
        progress: payload.progress,
        timestamp: toDate(payload.timestamp),
        finishedOn: toMaybeDate(payload.finishedOn),
        processedOn: toMaybeDate(payload.processedOn),
        failedReason: payload.failedReason,
        stacktrace: payload.stacktrace,
    }
}

/**
 * Format a job to return from the API.
 *
 * @param values A map of values composing the job.
 * @param status The job's status.
 */
export const formatJobFromMap = (values: Map<string, string>, status: string): ApiJob => {
    const rawData = values.get('data')
    const rawStacktrace = values.get('stacktrace')

    return {
        id: values.get('id') || '',
        name: values.get('name') || '',
        args: rawData ? JSON.parse(rawData).args : {},
        status,
        progress: toMaybeInt(values.get('progress')) || 0,
        timestamp: toMaybeDate(toMaybeInt(values.get('timestamp'))) || '',
        finishedOn: toMaybeDate(toMaybeInt(values.get('finishedOn'))),
        processedOn: toMaybeDate(toMaybeInt(values.get('processedOn'))),
        failedReason: values.get('failedReason') || null,
        stacktrace: rawStacktrace ? JSON.parse(rawStacktrace) : null,
    }
}
