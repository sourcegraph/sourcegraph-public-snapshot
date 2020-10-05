// We want to polyfill first.
import '../../shared/polyfills'

import React from 'react'
import { render } from 'react-dom'
import { AfterInstallPageContent } from '../after-install-page/AfterInstallPageContent'
import { ThemeWrapper } from '../ThemeWrapper'

// TODO dark theme support
// Share logic with webapp
document.body.classList.add('theme-light')

const AfterInstallPage: React.FunctionComponent = () => <ThemeWrapper>{() => <AfterInstallPageContent />}</ThemeWrapper>

render(<AfterInstallPage />, document.querySelector('#root'))
