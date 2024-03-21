import { type Writable, writable, type Updater, derived, type Readable } from 'svelte/store'

import { browser } from '$app/environment'

import { temporarySetting, type TemporarySettingStore } from './temporarySettings'

const LOCAL_STORAGE_THEME_KEY = 'sourcegraph-theme'

export enum Theme {
    Light = 'Light',
    Dark = 'Dark',
    System = 'System',
}

/**
 * The currently selected theme. Writing to this store will persist the theme
 * to temporary settings.
 */
export const theme: Writable<Theme> = (function () {
    let theme: Theme = ((browser && localStorage.getItem(LOCAL_STORAGE_THEME_KEY)) || Theme.System) as Theme
    let themeSettingStore: TemporarySettingStore<'user.themePreference'> | undefined

    const { subscribe } = writable(theme, set => {
        if (!themeSettingStore) {
            themeSettingStore = temporarySetting('user.themePreference')
        }

        return themeSettingStore.subscribe($themeSetting => {
            if (!$themeSetting.loading && $themeSetting.data) {
                switch ($themeSetting.data.toLowerCase()) {
                    case 'light': {
                        theme = Theme.Light
                        break
                    }
                    case 'dark': {
                        theme = Theme.Dark
                        break
                    }
                    case 'system': {
                        theme = Theme.System
                        break
                    }
                }
            }
            localStorage.setItem(LOCAL_STORAGE_THEME_KEY, theme)
            set(theme)
        })
    })

    function getThemeStringValue(theme: Theme): string {
        switch (theme) {
            case Theme.Light: {
                return 'light'
            }
            case Theme.Dark: {
                return 'dark'
            }
            case Theme.System: {
                return 'system'
            }
        }
    }

    return {
        subscribe,
        set: (value: Theme) => {
            themeSettingStore?.setValue(getThemeStringValue(value))
        },
        update: (updater: Updater<Theme>) => {
            themeSettingStore?.setValue(updater(theme))
        },
    }
})()

/**
 * This store returns true if the theme is set to light or if the user's system
 * preference is 'light'. The store listens to match media changes and updates
 * accordingly.
 */
export const isLightTheme = derived(theme, ($theme, set) => {
    if ($theme === Theme.System) {
        const matchMedia = window.matchMedia('(prefers-color-scheme: light)')
        set(matchMedia.matches)
        const listener = (event: MediaQueryListEventMap['change']): void => {
            set(event.matches)
        }
        matchMedia.addEventListener('change', listener)
        return () => matchMedia.removeEventListener('change', listener)
    }
    set($theme === Theme.Light)
    return
}) satisfies Readable<boolean>
