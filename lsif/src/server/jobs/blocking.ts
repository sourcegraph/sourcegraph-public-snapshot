import { Job } from 'bull'
import pTimeout from 'p-timeout'

/**
 * Wrap the given promise with a timeout, if the given timeout is a number.
 *
 * @param promise The promise to wrap.
 * @param maxWait The maximum time the promise can spend in-flight.
 */
const makePromise = <T>(promise: Promise<T>, maxWait: number): Promise<T> =>
    isNaN(maxWait) ? promise : pTimeout(promise, maxWait * 1000)

/**
 * Wait for the given job to resolve. The function resolves to true if the job
 * completed within the given timeout and false otherwise. If the job throws an
 * error, that error is thrown in-band. A NaN-valued max wait will block forever.
 *
 * @param job The job to block on.
 * @param maxWait The maximum time (in seconds) to wait for the promise to resolve.
 */
export const waitForJob = async (job: Job, maxWait: number): Promise<boolean> => {
    const promise = makePromise(job.finished(), maxWait)

    try {
        await promise
    } catch (error) {
        // Throw a job error, if one occurred.
        if (!error.message.includes('Promise timed out')) {
            throw error
        }

        return false
    }

    return true
}
