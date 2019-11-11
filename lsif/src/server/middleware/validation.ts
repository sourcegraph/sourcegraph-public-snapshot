import express from 'express'
import { query, ValidationChain, validationResult } from 'express-validator'

/**
 * Create a query string validator for a required non-empty string value.
 *
 * @param key The query string key.
 */
export const validateNonEmptyString = (key: string) =>
    query(key)
        .isString()
        .not()
        .isEmpty()

/**
 * Create a query string validator for a possibly empty string value.
 *
 * @param key The query string key.
 */
export const validateOptionalString = (key: string) =>
    query(key)
        .optional()
        .customSanitizer(value => value || '')

/**
 * Create a query string validator for a possibly empty boolean value.
 *
 * @param key The query string key.
 */
export const validateOptionalBoolean = (key: string) =>
    query(key)
        .optional()
        .isBoolean()
        .toBoolean()

/**
 * Create a query string validator for a possibly empty integer value.
 *
 * @param key The query string key.
 */
export const validateOptionalInt = (key: string) =>
    query(key)
        .optional()
        .isInt()
        .toInt()

/**
 * Middleware function used to apply a sequence of validators and then return
 * an unprocessable entity response with an error message if validation fails.
 */
export const validationMiddleware = (chains: ValidationChain[]) => async (
    req: express.Request,
    res: express.Response,
    next: express.NextFunction
): Promise<void> => {
    await Promise.all(chains.map(chain => chain.run(req)))

    var errors = validationResult(req)
    if (!errors.isEmpty()) {
        res.status(422).send({ errors: errors.mapped() })
        return
    }

    next()
}
