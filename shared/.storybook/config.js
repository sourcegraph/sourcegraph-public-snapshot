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
const globby = require('globby')
const path = require('path')

// Load CSS.
import '../../web/src/SourcegraphWebApp.scss'

// import '../src/components/Toggle.story'
// import '../src/actions/ActionItem.story'
// require('../src/components/Toggle.story')

async function main() {
  // Load all *.story.tsx files.
  if (require.context) {
    // In Webpack:
    const requireContext = require.context('../src', true, /\.story\.tsx?$/)
    for (const storyModule of requireContext.keys()) {
      requireContext(storyModule)
    }
  } else {
    // In Node (Jest):
    for (const storyModule of await globby(path.resolve(__dirname, '../src/**/*.story.tsx'))) {
      console.log('REQUIRE', storyModule)
      // require(storyModule)
    }
    require('../src/components/Toggle.story')
    // require('../src/actions/ActionItem.story')
    const y = await import('../src/components/Toggle.story')
    console.log('YYY', y)
    // await import('../src/actions/ActionItem.story')
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
