import type { Preview } from '@storybook/svelte'
import { initialize, mswLoader } from 'msw-storybook-addon'

import '../src/routes/styles.scss'

// Initialize MSW
initialize()

const preview: Preview = {
    parameters: {
        actions: { argTypesRegex: '^on[A-Z].*' },
        controls: {
            matchers: {
                color: /(background|color)$/i,
                date: /Date$/,
            },
        },
        darkMode: {
            stylePreview: true,
            darkClass: 'theme-dark',
            lightClass: 'theme-light',
        },
    },
    loaders: [mswLoader],
}

export default preview
