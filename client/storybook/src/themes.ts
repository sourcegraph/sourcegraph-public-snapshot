import { type ThemeVars, themes } from '@storybook/theming'
import openColor from 'open-color'

export const THEME_DARK_CLASS = 'theme-dark'
export const THEME_LIGHT_CLASS = 'theme-light'

// Themes use the colors from our webapp.
const common: Pick<
    ThemeVars,
    'colorPrimary' | 'colorSecondary' | 'brandTitle' | 'brandUrl' | 'brandImage' | 'fontBase' | 'fontCode'
> = {
    colorPrimary: openColor.blue[6],
    colorSecondary: openColor.blue[6],
    brandTitle: 'Sourcegraph Wildcard design system',
    brandUrl: 'https://sourcegraph.com',
    brandImage: '/img/wildcard-design-system.svg',
    fontBase:
        '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"',
    fontCode: 'sfmono-regular, consolas, menlo, dejavu sans mono, monospace',
}

export const themeDark: ThemeVars = {
    ...themes.dark,
    ...common,
    appBg: '#1c2736',
    appContentBg: '#151c28',
    appBorderColor: '#2b3750',
    barBg: '#0e121b',
    barTextColor: '#a2b0cd',
    textColor: '#f2f4f8',
    inputTextColor: '#ffffff',
}

export const themeLight: ThemeVars = {
    ...themes.light,
    ...common,
    appBg: '#fbfdff',
    textColor: '#2b3750',
    barTextColor: '#566e9f',
    inputTextColor: '#2b3750',
}
