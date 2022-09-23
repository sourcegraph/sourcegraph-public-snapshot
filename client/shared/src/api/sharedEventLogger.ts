// An interface that we can depent on in the shared package which is implemented
// by the event loggers.
export interface SharedEventLogger {
    log: (eventLabel: string, eventProperties?: any, publicArgument?: any) => void
}
