// We want to polyfill first.
import '../../shared/polyfills'

import React from 'react'

import { createRoot } from 'react-dom/client'

import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { WildcardThemeProvider } from '../../shared/components/WildcardThemeProvider'
import { AfterInstallPageContent } from '../after-install-page/AfterInstallPageContent'
import { ThemeWrapper } from '../ThemeWrapper'

setLinkComponent(AnchorLink)

const AfterInstallPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <ThemeWrapper>
        {({ isLightTheme }) => (
            <WildcardThemeProvider isBranded={true}>
                <AfterInstallPageContent isLightTheme={isLightTheme} />
            </WildcardThemeProvider>
        )}
    </ThemeWrapper>
)

const root = createRoot(document.querySelector('#root')!)

root.render(<AfterInstallPage />)
