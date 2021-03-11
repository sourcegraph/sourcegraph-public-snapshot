import { themes } from '@storybook/theming'
import openColor from 'open-color'
// @ts-ignore
import brandImage from '../ui/assets/img/wildcard-design-system.svg'

// Themes use the colors from our webapp.

/** @type {Partial<import('@storybook/theming').ThemeVars>} */
const common = {
  colorPrimary: openColor.blue[6],
  colorSecondary: openColor.blue[6],
  brandImage,
  brandTitle: 'Sourcegraph Wildcard design system',
  fontBase:
    '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"',
  fontCode: 'sfmono-regular, consolas, menlo, dejavu sans mono, monospace',
}

/** @type {import('@storybook/theming').ThemeVars} */
export const dark = {
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

/** @type {import('@storybook/theming').ThemeVars} */
export const light = {
  ...themes.light,
  ...common,
  appBg: '#fbfdff',
  textColor: '#2b3750',
  barTextColor: '#566e9f',
  inputTextColor: '#2b3750',
}
