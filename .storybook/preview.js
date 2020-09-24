import 'focus-visible'

import { configureActions } from '@storybook/addon-actions'
import { withConsole } from '@storybook/addon-console'
import { withKnobs } from '@storybook/addon-knobs'
import { addDecorator } from '@storybook/react'
import { setLinkComponent, AnchorLink } from '../shared/src/components/Link'
import { withDesign } from 'storybook-addon-designs'
import isChromatic from 'chromatic/isChromatic'

setLinkComponent(AnchorLink)

// Don't know why this type doesn't work, but this is the correct usage
// @ts-ignore
addDecorator(withKnobs)
addDecorator((storyFn, context) => withConsole()(storyFn)(context))
addDecorator(withDesign)

if (isChromatic()) {
  const style = document.createElement('style')
  style.innerHTML = `
      .monaco-editor .cursor {
        visibility: hidden !important;
      }
    `
  document.head.append(style)
}

configureActions({ depth: 100, limit: 20 })
