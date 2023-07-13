import type { Preview } from '@storybook/svelte'

import '../src/routes/styles.scss'

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
}

export default preview
