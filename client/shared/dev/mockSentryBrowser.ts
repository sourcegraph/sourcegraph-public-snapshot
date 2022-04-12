import * as Sentry from '@sentry/browser'

// eslint-disable-next-line no-var
declare var window: Window & typeof globalThis & { Sentry: typeof Sentry }

if ('Sentry' in window === false) {
    window.Sentry = Sentry
}
