// We want to polyfill first.
// prettier-ignore
import '../../app/util/polyfill'

import * as React from 'react'
import { render } from 'react-dom'
import { OptionsPage } from '../../app/components/options/OptionsPage'

const inject = () => {
    const injectDOM = document.createElement('div')
    injectDOM.id = 'sourcegraph-options-menu'
    document.body.appendChild(injectDOM)
    render(<OptionsPage />, injectDOM)
}

document.addEventListener('DOMContentLoaded', () => {
    inject()
})
