import React from 'react'
import { InstallBrowserExtensionAlert } from './InstallBrowserExtensionAlert'
import { ExternalServiceKind } from '../../../../shared/src/graphql/schema'
import { mount } from 'enzyme'
import { noop } from 'lodash'

describe('InstallBrowserExtensionAlert', () => {
    const serviceKinds = [
        ExternalServiceKind.GITHUB,
        ExternalServiceKind.GITLAB,
        ExternalServiceKind.PHABRICATOR,
        ExternalServiceKind.BITBUCKETSERVER,
        null,
    ] as const
    const integrationTypes = ['Chrome', 'non-Chrome', 'native integration'] as const
    for (const serviceKind of serviceKinds) {
        for (const integrationType of integrationTypes) {
            test(`${serviceKind ?? 'none'} (${integrationType})`, () => {
                expect(
                    mount(
                        <InstallBrowserExtensionAlert
                            isChrome={integrationType === 'Chrome'}
                            onAlertDismissed={noop}
                            codeHostIntegrationMessaging={
                                integrationType === 'native integration' ? 'native-integration' : 'browser-extension'
                            }
                            externalURLs={
                                serviceKind
                                    ? [
                                          {
                                              __typename: 'ExternalLink',
                                              url: '',
                                              serviceKind,
                                          },
                                      ]
                                    : []
                            }
                        />
                    )
                ).toMatchSnapshot()
            })
        }
    }
})
