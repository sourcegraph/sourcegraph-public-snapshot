import path from 'path'

import type { StorybookConfigVite } from '@storybook/builder-vite'
import type { StorybookConfig as ReactViteStorybookConfig } from '@storybook/react-vite'
import type { StorybookConfig } from '@storybook/types'

import { ROOT_PATH, STATIC_ASSETS_PATH } from '@sourcegraph/build-config'

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

const config: StorybookConfig & StorybookConfigVite & ReactViteStorybookConfig = {
    framework: '@storybook/react-vite',
    staticDirs: [path.resolve(__dirname, '../assets'), STATIC_ASSETS_PATH],
    stories: getStoriesGlob(),

    addons: [
        '@storybook/addon-actions',
        '@storybook/addon-designs',
        'storybook-dark-mode',
        '@storybook/addon-a11y',
        '@storybook/addon-toolbars',
        '@storybook/addon-controls',
        '@storybook/addon-storysource',
    ],

    core: {
        disableTelemetry: true,
        builder: '@storybook/builder-vite',
    },

    typescript: {
        check: false,
        reactDocgen: false,
    },

    viteFinal: (config, { configType }) => {
        config.build = {
            ...config.build,
            minify: false,

            // HACK(sqs): cssCodeSplit is needed to avoid `Failed to fetch dynamically imported
            // module: ...` errors where SourcegraphWebApp.scss's JavaScript stub file with the CSS
            // module class names is not emitted in the Storybook build. (It works in the dev
            // server.) This is not a perfect workaround as there are some incorrect global styles
            // being applied, but it's mostly fine (and any discrepancies are likely due to our
            // misuse of global CSS anyway).
            cssCodeSplit: false,
        }

        config.css = {
            ...config.css,
            modules: {
                ...config.css?.modules,
                localsConvention: 'camelCase',
            },
            preprocessorOptions: {
                scss: { includePaths: [path.resolve(ROOT_PATH, 'node_modules'), path.resolve(ROOT_PATH, 'client')] },
            },
        }

        config.optimizeDeps = {
            ...config.optimizeDeps,
            exclude: [
                ...(config.optimizeDeps?.exclude || []),
                // Work around an error on the Storybook dev server. See
                // https://github.com/vitejs/vite/issues/4245#issuecomment-879805637.
                'linguist-languages',
            ],
        }

        config.assetsInclude = ['**/*.yaml']
        config.resolve = {
            alias: {
                path: 'path-browserify',
                '@sourcegraph/shared/src/polyfills/vendor/eventSource': path.resolve(
                    __dirname,
                    'dummyEventSourcePolyfill'
                ),
            },
        }

        return config
    },
}

module.exports = config
