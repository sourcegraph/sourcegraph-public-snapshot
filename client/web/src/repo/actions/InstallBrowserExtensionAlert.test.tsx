import React from 'react'
import { InstallBrowserExtensionAlert, isFirefoxCampaignActive } from './InstallBrowserExtensionAlert'
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

        // TODO(tj): Remove this after the final date (December 31, 2020)
        describe(`FirefoxAddonAlert (${serviceType ?? 'unknown service type'})`, () => {
            test('displays Firefox addon alert before the final date', () => {
                expect(
                    mount(
                        <InstallBrowserExtensionAlert
                            showFirefoxAddonAlert={isFirefoxCampaignActive(new Date('December 23, 2020').getTime())}
                            onAlertDismissed={noop}
                            isChrome={true}
                            codeHostIntegrationMessaging="browser-extension"
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

            test('displays normal browser extension alert after the final date', () => {
                expect(
                    mount(
                        <InstallBrowserExtensionAlert
                            showFirefoxAddonAlert={isFirefoxCampaignActive(new Date('January 2, 2021').getTime())}
                            onAlertDismissed={noop}
                            isChrome={true}
                            codeHostIntegrationMessaging="browser-extension"
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
