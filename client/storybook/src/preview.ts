import 'focus-visible'
import { configureActions } from '@storybook/addon-actions'
import { withConsole } from '@storybook/addon-console'
import { DecoratorFunction } from '@storybook/addons'
import isChromatic from 'chromatic/isChromatic'
import { ReactElement } from 'react'
import { withDesign } from 'storybook-addon-designs'

import { setLinkComponent, AnchorLink } from '@sourcegraph/shared/src/components/Link'

import * as themes from './themes'

const withConsoleDecorator: DecoratorFunction<ReactElement> = (storyFn, context): ReactElement =>
    withConsole()(storyFn)(context)

export const decorators = [withDesign, withConsoleDecorator]

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

window.MonacoEnvironment = {
    getWorkerUrl(moduleId: string, label: string) {
        if (label === 'json') {
            return '/json.worker.bundle.js'
        }

        return '/editor.worker.bundle.js'
    },
}
