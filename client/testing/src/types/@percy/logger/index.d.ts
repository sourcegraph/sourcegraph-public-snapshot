declare module '@percy/logger' {
    type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'silent'

    interface Log {
        level: LogLevel
        message: string
    }

    interface LoggerFlags {
        verbose?: boolean
        quiet?: boolean
        silent?: boolean
    }

    export default function logger(debugLabel?: string): {
        info(message: string): void
        error(message: string): void
        warn(message: string): void
        debug(message: string): void
        deprecated(message: string): void
    }

    export function loglevel(level?: LogLevel, flags?: LoggerFlags): LogLevel
    export function format(message: string, debugLabel?: string, level?: LogLevel): string
    export function query(filter: (log: Log) => boolean): Log[]
}
