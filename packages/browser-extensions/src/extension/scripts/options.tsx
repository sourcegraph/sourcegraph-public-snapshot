// We want to polyfill first.
// prettier-ignore
import '../../config/polyfill'

import * as React from 'react'
import { render } from 'react-dom'
import storage from '../../browser/storage'
import { OptionsContainer, OptionsContainerProps } from '../../libs/options/OptionsContainer'
import { getConfigurableSettings, setConfigurabelSettings, setSourcegraphURL } from '../../libs/options/settings'
import { getAccessToken, setAccessToken } from '../../shared/auth/access_token'
import { createAccessToken, fetchAccessTokenIDs } from '../../shared/backend/auth'
import { fetchCurrentUser, fetchSite } from '../../shared/backend/server'
import { assertEnv } from '../envAssertion'

assertEnv('OPTIONS')

const inject = () => {
    const injectDOM = document.createElement('div')
    injectDOM.id = 'sourcegraph-options-menu'
    injectDOM.className = 'options'
    document.body.appendChild(injectDOM)

    const renderOptionsContainer = (sourcegraphURL: string) => {
        const props: OptionsContainerProps = {
            sourcegraphURL,

            ensureValidSite: fetchSite,
            fetchCurrentUser,

            setSourcegraphURL,
            getConfigurableSettings,
            setConfigurableSettings: setConfigurabelSettings,

            createAccessToken,
            getAccessToken,
            setAccessToken,
            fetchAccessTokenIDs,
        }

        render(<OptionsContainer {...props} />, injectDOM)
    }

    // storage.getSync(items => renderOptionsContainer(items.sourcegraphURL))
    storage.observeSync('sourcegraphURL').subscribe(url => {
        console.log('hello', url)
        renderOptionsContainer(url)
    })
}

document.addEventListener('DOMContentLoaded', () => {
    inject()
})
