import React from 'react'
import { InstallBrowserExtensionAlert } from './InstallBrowserExtensionAlert'
import { mount } from 'enzyme'
import { noop } from 'lodash'

describe('InstallBrowserExtensionAlert', () => {
    test('GitHub', () => {
        expect(
            mount(
                <InstallBrowserExtensionAlert
                    onAlertDismissed={noop}
                    externalURLs={[
                        {
                            __typename: 'ExternalLink',
                            url: 'https://github.com/sourcegraph/sourcegraph',
                            serviceType: 'github',
                        },
                    ]}
                />
            )
        ).toMatchSnapshot()
    })

    test('GitLab', () => {
        expect(
            mount(
                <InstallBrowserExtensionAlert
                    onAlertDismissed={noop}
                    externalURLs={[
                        {
                            __typename: 'ExternalLink',
                            url: 'https://gitlab.com/rluna-open-source/code-management/sourcegraph/sourcegraph',
                            serviceType: 'gitlab',
                        },
                    ]}
                />
            )
        ).toMatchSnapshot()
    })
})
