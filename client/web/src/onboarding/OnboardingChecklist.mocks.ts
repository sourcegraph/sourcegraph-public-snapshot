import type { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import type { SiteConfigResult } from '../graphql-operations'

import { SITE_CONFIG_QUERY } from './queries'

export const emptySiteConfig = (): MockedResponse<SiteConfigResult> => ({
    request: {
        query: getDocumentNode(SITE_CONFIG_QUERY),
    },
    result: {
        data: {
            site: {
                configuration: {
                    id: 1,
                    effectiveContents:
                        '{"licenseKey":"","externalURL":"","email.smtp":{"host":""},"auth.providers":[]}',
                    licenseInfo: {
                        tags: [],
                        userCount: 10,
                        expiresAt: '',
                    },
                },
            },
            externalServices: {
                nodes: [],
            },
        },
    },
})

export const completeSiteConfig = (): MockedResponse<SiteConfigResult> => ({
    request: {
        query: getDocumentNode(SITE_CONFIG_QUERY),
    },
    result: {
        data: {
            site: {
                configuration: {
                    id: 1,
                    effectiveContents:
                        '{"licenseKey":"mockedLicenseKey","externalURL":"sourcegraph.com","email.smtp":{"host":"127.0.0.1"},"auth.providers":[{"type": "builtin"}, {"type": "github"}]}',
                    licenseInfo: {
                        tags: ['enterprise'],
                        userCount: 10,
                        expiresAt: '2028-01-01',
                    },
                },
            },
            externalServices: {
                nodes: [
                    {
                        id: '1',
                        displayName: 'GitHub',
                    },
                ],
            },
        },
    },
})

export const incompleteSiteConfig = (): MockedResponse<SiteConfigResult> => ({
    request: {
        query: getDocumentNode(SITE_CONFIG_QUERY),
    },
    result: {
        data: {
            site: {
                configuration: {
                    id: 1,
                    effectiveContents:
                        '{"licenseKey":"mockedLicenseKey","externalURL":"","email.smtp":{"host":"127.0.0.1"},"auth.providers":[{"type": "builtin"}]}',
                    licenseInfo: {
                        tags: ['enterprise'],
                        userCount: 10,
                        expiresAt: '2028-01-01',
                    },
                },
            },
            externalServices: {
                nodes: [
                    {
                        id: '1',
                        displayName: 'GitHub',
                    },
                ],
            },
        },
    },
})
