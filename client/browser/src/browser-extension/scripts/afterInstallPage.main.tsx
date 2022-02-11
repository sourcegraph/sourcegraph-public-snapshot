// We want to polyfill first.
import '../../shared/polyfills'

import React from 'react'
import { render } from 'react-dom'

import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { WildcardThemeProvider } from '../../shared/components/WildcardThemeProvider'
import { AfterInstallPageContent } from '../after-install-page/AfterInstallPageContent'
import { ThemeWrapper } from '../ThemeWrapper'

setLinkComponent(AnchorLink)

const AfterInstallPage: React.FunctionComponent = () => (
    <ThemeWrapper>
        {({ isLightTheme }) => (
            <WildcardThemeProvider isBranded={true}>
                <AfterInstallPageContent isLightTheme={isLightTheme} />
            </WildcardThemeProvider>
        )}
    </ThemeWrapper>
)

render(<AfterInstallPage />, document.querySelector('#root'))
