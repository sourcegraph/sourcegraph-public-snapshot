// We want to polyfill first.
// prettier-ignore
import '../../config/polyfill'

import * as React from 'react'
import { render } from 'react-dom'
import storage from '../../browser/storage'
import { OptionsDashboard } from '../../shared/components/options/OptionsDashboard'
import { assertEnv } from '../envAssertion'

assertEnv('OPTIONS')

const inject = () => {
    const injectDOM = document.createElement('div')
    injectDOM.id = 'sourcegraph-options-menu'
    injectDOM.className = 'options'
    document.body.appendChild(injectDOM)

    storage.getSync(items => {
        render(<OptionsDashboard />, injectDOM)
    })
}

document.addEventListener('DOMContentLoaded', () => {
    inject()
})
