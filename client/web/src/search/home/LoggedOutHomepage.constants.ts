import { DynamicWebFont } from './DynamicWebFonts'

/**
 * Source Sans Pro fonts: https://fonts.google.com/specimen/Source+Sans+Pro
 * Two families required for the `LoggedOutHomepage` UI: Regular 400 and Bold 700.
 *
 * Assets are downloaded from the Google generated link. Only latin glyphs are included.
 * https://fonts.googleapis.com/css2?family=Source+Sans+Pro:wght@400;700&display=swap
 */
export const fonts: DynamicWebFont[] = [
    {
        family: 'Source Sans Pro',
        weight: '400',
        source: `url(${JSON.stringify(
            new URL('./SourceSansPro-Regular.woff2', import.meta.url).toString()
        )}) format('woff2')`,
        style: 'normal',
        unicodeRange:
            'U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+2000-206F, U+2074, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD',
    },
    {
        family: 'Source Sans Pro',
        weight: '700',
        source: `url(${JSON.stringify(
            new URL('./SourceSansPro-Bold.woff2', import.meta.url).toString()
        )}) format('woff2')`,
        style: 'normal',
        unicodeRange:
            'U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+2000-206F, U+2074, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD',
    },
]
