import { IsProductionVersion } from './util'

/**
 * The "browser action" is the name for the icon/button which is shown in the
 * browser UI (in the toolbar).
 *
 * The browser action icon is the icon image, provided as PNGs of different
 * sizes for different resolutions, and which can be dynamically changed based
 * on the state of the browser extension.
 */

/** State of the Sourcegraph browser extension, as represented by the browser
 * action icon. */
export type BrowserActionIconState = 'active' | 'active-with-alert' | 'inactive'
interface BrowserActionIconPaths {
    '32': string
    '48': string
    '128': string
}

const browserActionIconPaths: Record<BrowserActionIconState, BrowserActionIconPaths> = IsProductionVersion
    ? {
          active: {
              '32': 'img/icon-32.png',
              '48': 'img/icon-48.png',
              '128': 'img/icon-128.png',
          },
          'active-with-alert': {
              '32': 'img/icon-active-with-alert-32.png',
              '48': 'img/icon-active-with-alert-48.png',
              '128': 'img/icon-active-with-alert-128.png',
          },
          inactive: {
              '32': 'img/icon-inactive-32.png',
              '48': 'img/icon-inactive-48.png',
              '128': 'img/icon-inactive-128.png',
          },
      }
    : {
          active: {
              '32': 'img/dev/icon-32.png',
              '48': 'img/dev/icon-48.png',
              '128': 'img/dev/icon-128.png',
          },
          'active-with-alert': {
              '32': 'img/dev/icon-active-with-alert-32.png',
              '48': 'img/dev/icon-active-with-alert-48.png',
              '128': 'img/dev/icon-active-with-alert-128.png',
          },
          inactive: {
              '32': 'img/dev/icon-inactive-32.png',
              '48': 'img/dev/icon-inactive-48.png',
              '128': 'img/dev/icon-inactive-128.png',
          },
      }

/**
 * Update the browser action icon to the given state.
 */
export function setBrowserActionIconState(iconState: BrowserActionIconState): void {
    const iconPaths = browserActionIconPaths[iconState]
    console.log('Setting icons to', iconPaths)
    browser.browserAction.setIcon({ path: iconPaths }).catch(error => {
        console.error(error)
    })
}
