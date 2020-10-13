import React from 'react'
import { InstallBrowserExtensionAlert } from './InstallBrowserExtensionAlert'
import { mount } from 'enzyme'
import { noop } from 'lodash'

describe('InstallBrowserExtensionAlert', () => {
    const serviceTypes = ['github', 'gitlab', 'phabricator', 'bitbucketServer'] as const
    const integrationTypes = ['Chrome', 'non-Chrome', 'native integration'] as const
    for (const serviceType of serviceTypes) {
        for (const integrationType of integrationTypes) {
            test(`${serviceType} (${integrationType})`, () => {
                expect(
                    mount(
                        <InstallBrowserExtensionAlert
                            isChrome={integrationType === 'Chrome'}
                            onAlertDismissed={noop}
                            isNativeIntegrationEnabled={integrationType === 'native integration'}
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
