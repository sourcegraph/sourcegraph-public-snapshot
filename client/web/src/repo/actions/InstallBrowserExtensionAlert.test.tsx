import { render } from '@testing-library/react'
import { noop } from 'lodash'
import React from 'react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'

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
                    render(
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
                    ).asFragment()
                ).toMatchSnapshot()
            })
        }
    }
})
