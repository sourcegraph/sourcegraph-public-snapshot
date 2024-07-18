import type { Preview } from '@storybook/svelte'
import { initialize, mswLoader } from 'msw-storybook-addon'

// Global imports kept in sync with routes/+layout.svelte
import '../src/routes/styles.scss'
import '@fontsource-variable/roboto-mono'
import '@fontsource-variable/inter'

// Initialize MSW
initialize()

const preview: Preview = {
    parameters: {
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
