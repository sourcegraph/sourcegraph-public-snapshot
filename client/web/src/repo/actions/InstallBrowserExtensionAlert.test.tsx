import React from 'react'
import { InstallBrowserExtensionAlert } from './InstallBrowserExtensionAlert'
import { mount } from 'enzyme'
import { noop } from 'lodash'

describe('InstallBrowserExtensionAlert', () => {
    const serviceTypes = ['github', 'gitlab', 'phabricator', 'bitbucketServer'] as const
    const browsers = ['Chrome', 'non-Chrome'] as const
    for (const serviceType of serviceTypes) {
        for (const browser of browsers) {
            test(`${serviceType} (${browser})`, () => {
                expect(
                    mount(
                        <InstallBrowserExtensionAlert
                            isChrome={browser === 'Chrome'}
                            onAlertDismissed={noop}
                            externalURLs={[
                                {
                                    __typename: 'ExternalLink',
                                    url: '',
                                    serviceType,
                                },
                            ]}
                        />
                    )
                ).toMatchSnapshot()
            })
        }
    }
})
