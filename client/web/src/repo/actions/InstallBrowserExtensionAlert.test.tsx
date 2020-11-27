import React from 'react'
import { FirefoxAddonAlert, InstallBrowserExtensionAlert } from './InstallBrowserExtensionAlert'
import { mount } from 'enzyme'
import { noop } from 'lodash'

describe('InstallBrowserExtensionAlert', () => {
    const serviceTypes = ['github', 'gitlab', 'phabricator', 'bitbucketServer', null] as const
    const integrationTypes = ['Chrome', 'non-Chrome', 'native integration'] as const
    for (const serviceType of serviceTypes) {
        for (const integrationType of integrationTypes) {
            test(`${serviceType ?? 'none'} (${integrationType})`, () => {
                expect(
                    mount(
                        <InstallBrowserExtensionAlert
                            isChrome={integrationType === 'Chrome'}
                            onAlertDismissed={noop}
                            codeHostIntegrationMessaging={
                                integrationType === 'native integration' ? 'native-integration' : 'browser-extension'
                            }
                            externalURLs={
                                serviceType
                                    ? [
                                          {
                                              __typename: 'ExternalLink',
                                              url: '',
                                              serviceType,
                                          },
                                      ]
                                    : []
                            }
                        />
                    )
                ).toMatchSnapshot()
            })
        }

        // TODO(tj): Remove this after the final date (December 21, 2020)
        describe(`FirefoxAddonAlert (${serviceType ?? 'unknown service type'})`, () => {
            test('displays alert before the final date', () => {
                expect(
                    mount(
                        <FirefoxAddonAlert
                            now={() => new Date('December 10, 2020')}
                            onAlertDismissed={noop}
                            externalURLs={
                                serviceType
                                    ? [
                                          {
                                              __typename: 'ExternalLink',
                                              url: '',
                                              serviceType,
                                          },
                                      ]
                                    : []
                            }
                        />
                    )
                ).toMatchSnapshot()
            })

            test('does not display test after the final date', () => {
                expect(
                    mount(
                        <FirefoxAddonAlert
                            now={() => new Date('December 22, 2020')}
                            onAlertDismissed={noop}
                            externalURLs={
                                serviceType
                                    ? [
                                          {
                                              __typename: 'ExternalLink',
                                              url: '',
                                              serviceType,
                                          },
                                      ]
                                    : []
                            }
                        />
                    )
                ).toMatchSnapshot()
            })
        })
    }
})
