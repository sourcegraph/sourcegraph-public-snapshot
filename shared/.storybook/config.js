// @ts-check

import { configure, addDecorator } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'
import { withOptions } from '@storybook/addon-options'
import { themes } from '@storybook/components'
import { configureActions } from '@storybook/addon-actions'
import '@storybook/addon-console'
import { withConsole } from '@storybook/addon-console'
import { withNotes } from '@storybook/addon-notes'
import '../../web/src/SourcegraphWebApp.scss'

addDecorator(withNotes())
addDecorator((storyFn, context) => withConsole()(storyFn)(context))
addDecorator(withOptions({ theme: themes.dark }))
addDecorator(withInfo({ header: false, propTables: false }))

configureActions({
  depth: 100,
  limit: 20,
})

async function main() {
  // Load all *.story.tsx files.
  const requireContext = require.context('../src', true, /\.story\.tsx?$/)
  for (const storyModule of requireContext.keys()) {
    requireContext(storyModule)
  }

  // Configure storybooks.
  configure(() => void 0, module)
}
main().catch(err => console.error(err))
