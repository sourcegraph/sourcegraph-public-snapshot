import winston from 'winston'

/**
 * TODO
 */
const LOG_LEVEL = 'debug' || process.env.LOG_LEVEL || 'info' // TODO - temporary

/**
 * TODO
 */
export const logger = winston.createLogger({
    level: LOG_LEVEL,
    format: winston.format.json(),
    defaultMeta: {
        service: 'lsif-server',
    },
    transports: [
        new winston.transports.Console({
            format: winston.format.simple(),
        }),
    ],
})
