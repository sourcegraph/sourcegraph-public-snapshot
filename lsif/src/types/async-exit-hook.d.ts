declare module 'async-exit-hook'

declare function exitHook(f: () => Promise<void> | void): void
