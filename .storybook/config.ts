import { configureActions } from '@storybook/addon-actions'
// @ts-ignore
import { withConsole } from '@storybook/addon-console'
import { withInfo } from '@storybook/addon-info'
import { withKnobs } from '@storybook/addon-knobs'
import { addDecorator, addParameters, configure } from '@storybook/react'
import { themes } from '@storybook/theming'

import './styles'

async function main(): Promise<void> {
    // Webpack provides require.context. TODO: If this is run in Jest in the future, we'll need to
    // use babel-plugin-require-context-hook.
    const requireContexts = [
        require.context('../shared', true, /\.story\.tsx?$/),
        require.context('../client/browser', true, /\.story\.tsx?$/),
        require.context('../web', true, /\.story\.tsx?$/),
    ]
    for (const requireContext of requireContexts) {
        for (const storyModule of requireContext.keys()) {
            requireContext(storyModule)
        }
    }

    // Configure storybooks.
    configure(() => {
        addDecorator((storyFn, context) => withKnobs(storyFn, context))
        addDecorator((storyFn, context) => withConsole()(storyFn)(context))
        addParameters({ theme: themes.dark })
        addDecorator(withInfo({ header: false, propTables: false }))

        configureActions({
            depth: 100,
            limit: 20,
        })
    }, module)
}
main().catch(err => console.error(err))
