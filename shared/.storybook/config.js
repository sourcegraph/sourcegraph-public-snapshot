// @ts-check

import { configure, addDecorator } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'
import { withOptions } from '@storybook/addon-options'
// @ts-ignore
import { themes } from '@storybook/components'
import { configureActions } from '@storybook/addon-actions'
import '@storybook/addon-console'
// @ts-ignore
import { withConsole } from '@storybook/addon-console'
import { withNotes } from '@storybook/addon-notes'

// Load CSS.
import '../../web/src/SourcegraphWebApp.scss'

async function main() {
  // Load all *.story.tsx files.
  if (require.context) {
    // In Webpack:
    const requireContext = require.context('../src', true, /\.story\.tsx?$/)
    for (const storyModule of requireContext.keys()) {
      requireContext(storyModule)
    }
  } else {
    throw new Error('running storybooks outside of Webpack is not yet supported')
  }

  // Configure storybooks.
  configure(() => {
    addDecorator(withNotes())
    addDecorator((storyFn, context) => withConsole()(storyFn)(context))
    // @ts-ignore
    addDecorator(withOptions({ theme: themes.dark }))
    addDecorator(withInfo({ header: false, propTables: false }))

    configureActions({
      depth: 100,
      limit: 20,
    })
  }, module)
}
main().catch(err => console.error(err))
