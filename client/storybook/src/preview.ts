import 'focus-visible'

import type { ReactElement } from 'react'

import { configureActions } from '@storybook/addon-actions'
import { withConsole } from '@storybook/addon-console'
import type { DecoratorFn, Parameters } from '@storybook/react'

import { setLinkComponent, AnchorLink } from '@sourcegraph/wildcard'

import { themeDark, themeLight, THEME_DARK_CLASS, THEME_LIGHT_CLASS } from './themes'

const withConsoleDecorator: DecoratorFn = (storyFunc, context): ReactElement => withConsole()(storyFunc)(context)

export const decorators = [withConsoleDecorator].filter(Boolean)

export const parameters: Parameters = {
    layout: 'fullscreen',
    options: {
        storySort: {
            order: ['wildcard', 'shared', 'branded', '*'],
            method: 'alphabetical',
        },
    },
    darkMode: {
        stylePreview: true,
        lightClass: THEME_LIGHT_CLASS,
        darkClass: THEME_DARK_CLASS,
        light: themeLight,
        dark: themeDark,
    },
}

configureActions({ depth: 100, limit: 20 })

setLinkComponent(AnchorLink)

declare global {
    interface Window {
        STORYBOOK_ENV?: string
    }
}

/**
 * Since we do not use `storiesOf` API, this env variable is not set by `@storybook/react` anymore.
 * The `withConsole` decorator relies on this env variable so we set it manually here.
 */
if (!window.STORYBOOK_ENV) {
    window.STORYBOOK_ENV = 'react'
}
