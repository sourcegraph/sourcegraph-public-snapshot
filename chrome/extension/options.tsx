// We want to polyfill first.
// prettier-ignore
import '../../app/util/polyfill'

import * as React from 'react'
import { render } from 'react-dom'
import { OptionsDashboard } from '../../app/components/options/OptionsDashboard'
import storage from '../../extension/storage'

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
