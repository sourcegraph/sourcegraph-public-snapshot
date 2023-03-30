type LogFunc = (msg: string) => void

// Logger is a singleton class that allows us to log messages to the Debug tab.
// To use it, call Logger.getInstance().log('message') and make sure that
// setLogFunc() has been called upstream.
//
/* eslint no-empty-function: ["error", { "allow": ["constructors"] }]*/
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

    public setLogFunc(logFunc: LogFunc): void {
        Logger.logFunc = logFunc
    }

    public log(message: string): void {
        Logger.logFunc?.(message)
    }

    public logJSON(
        message: any,
        replacer?: (this: any, key: string, value: any) => any,
        space?: string | number
    ): void {
        Logger.logFunc?.(JSON.stringify(message, replacer, 2))
    }
}
