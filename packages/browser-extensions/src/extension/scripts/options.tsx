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
import { OptionsDashboard } from '../../shared/components/options/OptionsDashboard'
import { featureFlags } from '../../shared/util/featureFlags'
import { assertEnv } from '../envAssertion'

assertEnv('OPTIONS')

const inject = async () => {
    const injectDOM = document.createElement('div')
    injectDOM.id = 'sourcegraph-options-menu'
    injectDOM.className = 'options'
    document.body.appendChild(injectDOM)

    if (await featureFlags.isEnabled('simpleOptionsMenu')) {
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

        storage.observeSync('sourcegraphURL').subscribe(url => {
            renderOptionsContainer(url)
        })
    } else {
        storage.getSync(() => {
            render(<OptionsDashboard />, injectDOM)
        })
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    await inject()
})
