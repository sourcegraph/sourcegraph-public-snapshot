import type { StorybookConfig } from '@storybook/sveltekit'

const config: StorybookConfig = {
    stories: ['../src/**/*.mdx', '../src/**/*.stories.svelte'],
    addons: ['@storybook/addon-essentials', '@storybook/addon-svelte-csf', 'storybook-dark-mode'],
    framework: {
        name: '@storybook/sveltekit',
        options: {},
    },
    staticDirs: ['../static'],
    docs: {
        autodocs: false,
    },
    core: {
        disableTelemetry: true,
    },
}
export default config
