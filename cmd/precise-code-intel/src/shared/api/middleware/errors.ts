import express from 'express'
import { Logger } from 'winston'

interface ErrorResponse {
    message: string
}

export interface ApiError {
    message: string
    status?: number
}

export const isApiError = (val: unknown): val is ApiError => typeof val === 'object' && !!val && 'message' in val

/**
 * Middleware function used to convert uncaught exceptions into 500 responses.
 *
 * @param logger The logger instance.
 */
export const errorHandler = (
    logger: Logger
): ((
    error: unknown,
    req: express.Request,
    res: express.Response<ErrorResponse>,
    next: express.NextFunction
) => void) => (
    error: unknown,
    req: express.Request,
    res: express.Response<ErrorResponse>,
    // Express uses argument length to distinguish middleware and error handlers
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    next: express.NextFunction
): void => {
    const status = (isApiError(error) && error.status) || 500
    const message = (isApiError(error) && error.message) || 'Unknown error'

    if (status === 500) {
        logger.error('uncaught exception', { error })
    }

    if (!res.headersSent) {
        res.status(status).send({ message })
    }
}
