type LogFunc = (msg: string) => Promise<void>

// Logger is a singleton class that allows us to log messages to the Debug tab.
// To use it, call Logger.getInstance().log('message') and make sure that
// setLogFunc() has been called upstream.
export class Logger {
    private static logFunc: LogFunc | undefined
    private static instance: Logger

    private constructor() {} // eslint-disable-line @typescript-eslint/no-empty-function

    public static getInstance(): Logger {
        if (!Logger.instance) {
            Logger.instance = new Logger()
        }
        return Logger.instance
    }

    public setLogFunc(logFunc: LogFunc): void {
        Logger.logFunc = logFunc
    }

    public async log(message: string): Promise<void> {
        return Logger.logFunc?.(message)
    }

    // logJson has the same signature as JSON.stringify
    public async logJSON(
        message: any,
        replacer?: (this: any, key: string, value: any) => any,
        space?: string | number
    ): Promise<void> {
        return Logger.logFunc?.(JSON.stringify(message, replacer, 2))
    }
}
