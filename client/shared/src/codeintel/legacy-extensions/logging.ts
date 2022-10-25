import { inspect } from 'util'

export type LogLevel = 'error' | 'warn' | 'info' | 'log'
export type Logger = Record<LogLevel, (...values: any[]) => void>

/** Logger implementation that does nothing. */
export class NoopLogger {
    public error(): void {
        /* no-op */
    }

    public warn(): void {
        /* no-op */
    }

    public info(): void {
        /* no-op */
    }

    public log(): void {
        /* no-op */
    }
}

/** Logger that formats the logged values and removes any auth info in URLs. */
export class RedactingLogger {
    constructor(private logger: Logger) {}

    public error(...values: any[]): void {
        this.logger.error(...values.map(redact))
    }

    public warn(...values: any[]): void {
        this.logger.warn(...values.map(redact))
    }

    public info(...values: any[]): void {
        this.logger.info(...values.map(redact))
    }

    public log(...values: any[]): void {
        this.logger.log(...values.map(redact))
    }
}

/** Removes auth info from URLs. */
function redact(value: any): string {
    const stringValue = typeof value === 'string' ? value : inspect(value, { depth: Infinity })

    return stringValue.replace(/(https?:\/\/)[^/@]+@([^\s$]+)/g, '$1$2')
}
