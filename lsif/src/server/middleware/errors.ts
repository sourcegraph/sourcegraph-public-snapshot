import express from 'express'
import { Logger } from 'winston'
import { pipeline as _pipeline } from 'stream'

/**
 * Middleware function used to convert uncaught exceptions into 500 responses.
 *
 * @param logger The logger instance.
 */
export const errorHandler = (
    logger: Logger
): ((error: any, req: express.Request, res: express.Response, next: express.NextFunction) => void) => (
    error: any,
    req: express.Request,
    res: express.Response,
    next: express.NextFunction
): void => {
    if (!error || !error.status) {
        // Only log errors that don't have a status attached
        logger.error('uncaught exception', { error })
    }

    if (!res.headersSent) {
        res.status((error && error.status) || 500).send({ message: (error && error.message) || 'Unknown error' })
    }
}
