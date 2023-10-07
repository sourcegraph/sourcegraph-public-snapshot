import path from 'path'

import type { StorybookConfigVite } from '@storybook/builder-vite'
import type { StorybookConfig } from '@storybook/core-common'

import { ROOT_PATH, STATIC_ASSETS_PATH, getEnvironmentBoolean } from '@sourcegraph/build-config'

import { ENVIRONMENT_CONFIG } from './environment-config'

const getStoriesGlob = (): string[] => {
    if (ENVIRONMENT_CONFIG.STORIES_GLOB) {
        return [path.resolve(ROOT_PATH, ENVIRONMENT_CONFIG.STORIES_GLOB)]
    }

    // Due to an issue with constant recompiling (https://github.com/storybookjs/storybook/issues/14342)
    // we need to make the globs more specific (`(web|shared..)` also doesn't work). Once the above issue
    // is fixed, this can be removed and watched for `client/**/*.story.tsx` again.
    const directoriesWithStories = ['branded', 'browser', 'jetbrains/webview', 'shared', 'web', 'wildcard']
    const storiesGlobs = directoriesWithStories.map(packageDirectory =>
        path.resolve(ROOT_PATH, `client/${packageDirectory}/src/**/*.story.tsx`)
    )

    return [...storiesGlobs]
}

const isStoryshotsEnvironment = globalThis.navigator?.userAgent?.match?.('jsdom')

const config: StorybookConfig & StorybookConfigVite = {
    framework: '@storybook/react',
    staticDirs: [path.resolve(__dirname, '../assets'), STATIC_ASSETS_PATH],
    stories: getStoriesGlob(),
    addons: [
        '@storybook/addon-actions',
        'storybook-addon-designs',
        'storybook-dark-mode',
        '@storybook/addon-a11y',
        '@storybook/addon-toolbars',
        '@storybook/addon-docs',
        '@storybook/addon-controls',
        {
            name: '@storybook/addon-storysource',
            options: {
                rule: {
                    test: /\.story\.tsx?$/,
                },
                sourceLoaderOptions: {
                    injectStoryParameters: false,
                    prettierConfig: { printWidth: 80, singleQuote: false },
                },
            },
        },
    ],

    core: {
        disableTelemetry: true,
        builder: '@storybook/builder-vite',
    },

    features: {
        // Explicitly disable the deprecated, not used postCSS support,
        // so no warning is rendered on each start of storybook.
        postcss: false,
        // Storyshots is not currently compatible with the v7 store.
        // https://github.com/storybookjs/storybook/blob/next/MIGRATION.md#storyshots-compatibility-in-the-v7-store
        storyStoreV7: !isStoryshotsEnvironment,
        babelModeV7: !isStoryshotsEnvironment,
    },

    typescript: {
        check: false,
        reactDocgen: false,
    },

    viteFinal: config => {
        config.define = { ...config.define, CHROMATIC: getEnvironmentBoolean('CHROMATIC') }
        return config
    },
}

module.exports = config
