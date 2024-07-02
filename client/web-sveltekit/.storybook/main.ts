import type { StorybookConfig } from '@storybook/sveltekit'
import type { InlineConfig, Plugin } from 'vite'

// See https://github.com/storybookjs/storybook/issues/20562
const workaroundSvelteDocgenPluginConflictWithUnpluginIcons = (config: InlineConfig) => {
    if (!config.plugins) return config

    const [_internalPlugins, ...userPlugins] = config.plugins as Plugin[]
    const docgenPlugin = userPlugins.find(plugin => plugin.name === 'storybook:svelte-docgen-plugin')
    if (docgenPlugin) {
        const origTransform = docgenPlugin.transform
        const newTransform: typeof origTransform = (code, id, options) => {
            if (id.startsWith('~icons/')) {
                return
            }
            return (origTransform as Function)?.call(docgenPlugin, code, id, options)
        }
        docgenPlugin.transform = newTransform
        docgenPlugin.enforce = 'post'
    }
    return config
}

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
    viteFinal(config) {
        return workaroundSvelteDocgenPluginConflictWithUnpluginIcons(config)
    },
    core: {
        disableTelemetry: true,
    },
}
export default config
