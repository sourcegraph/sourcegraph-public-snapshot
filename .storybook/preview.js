import 'focus-visible'

import { configureActions } from '@storybook/addon-actions'
import { withConsole } from '@storybook/addon-console'
import { setLinkComponent, AnchorLink } from '../client/shared/src/components/Link'
import { withDesign } from 'storybook-addon-designs'
import isChromatic from 'chromatic/isChromatic'
import * as themes from './themes'

export const decorators = [withDesign, (storyFn, context) => withConsole()(storyFn)(context)]

export const parameters = {
  darkMode: {
    stylePreview: true,
    darkClass: 'theme-dark',
    lightClass: 'theme-light',
    light: themes.light,
    dark: themes.dark,
  },
}

configureActions({ depth: 100, limit: 20 })

setLinkComponent(AnchorLink)

// Default to light theme for Chromatic and "Open canvas in new tab" button.
// addon-dark-mode will override this if it's running.
if (!document.body.classList.contains('theme-dark')) {
  document.body.classList.add('theme-light')
}

if (isChromatic()) {
  const style = document.createElement('style')
  style.innerHTML = `
      .monaco-editor .cursor {
        visibility: hidden !important;
      }
    `
  document.head.append(style)
}

// @ts-ignore
window.MonacoEnvironment = {
  getWorkerUrl(_, label) {
    if (label === 'json') {
      return '/json.worker.bundle.js'
    }
    return '/editor.worker.bundle.js'
  },
}
