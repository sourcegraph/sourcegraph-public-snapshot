import { mount } from 'enzyme'
import { noop } from 'lodash'
import React from 'react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql/schema'

import { InstallBrowserExtensionAlert } from './InstallBrowserExtensionAlert'

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
