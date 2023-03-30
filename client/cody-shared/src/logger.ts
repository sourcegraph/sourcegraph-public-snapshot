interface LogFunc {
    (msg: string): void
}

// Logger is a singleton class that allows us to log messages to the Debug tab.
// To use it, call Logger.getInstance().log('message') and make sure that
// setLogFunc() has been called upstream.
export class Logger {
    private static logFunc: LogFunc | undefined
    private static instance: Logger

    private constructor() {}

    public static getInstance(): Logger {
        if (!Logger.instance) {
            Logger.instance = new Logger()
        }
        return Logger.instance
    }

    public setLogFunc(logFunc: LogFunc) {
        Logger.logFunc = logFunc
    }

    log(message: string) {
        Logger.logFunc?.(message)
    }

    logJSON(message: any, replacer?: (this: any, key: string, value: any) => any, space?: string | number) {
        Logger.logFunc?.(JSON.stringify(message, replacer, 2))
    }
}
