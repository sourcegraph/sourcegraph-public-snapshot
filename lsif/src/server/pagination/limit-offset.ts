import express from 'express'

/**
 * Extract a page limit from the request query string.
 *
 * @param req The HTTP request.
 * @param defaultLimit The limit to use if one is not supplied.
 */
export function extractLimit(req: express.Request, defaultLimit: number): number {
    return parseInt(req.query.limit, 10) || defaultLimit
}

/**
 * Extract a page limit and offset from the request query string.
 *
 * @param req The HTTP request.
 * @param defaultLimit The limit to use if one is not supplied.
 */
export function limitOffset(req: express.Request, defaultLimit: number): { limit: number; offset: number } {
    return {
        limit: parseInt(req.query.limit, 10) || defaultLimit,
        offset: parseInt(req.query.offset, 10) || 0,
    }
}
