import { shuffle } from 'lodash'
import React from 'react'
import { Logo1, Logo2, Logo3 } from './logos'

// Shuffle logos because we love all of them infinitely. :)
const LOGOS: {
    component: React.ComponentType<{ className: string; isLightTheme: boolean }>
    className: string
}[] = shuffle([
    {
        component: Logo1,
        className: 'welcome-main-page-logos__logo-1 mr-3',
    },
    {
        component: Logo2,
        className: 'welcome-main-page-logos__logo-2 mr-3',
    },
    {
        component: Logo3,
        className: 'welcome-main-page-logos__logo-3 mr-3',
    },
])

/**
 * The logos for the welcome main page.
 */
export const WelcomeMainPageLogos: React.FunctionComponent<{ isLightTheme: boolean }> = ({ isLightTheme }) => (
    <>
        {LOGOS.map(({ component: C, className }, i) => (
            <C key={i} className={`welcome-main-page-logos__logo ${className}`} isLightTheme={isLightTheme} />
        ))}
    </>
)
