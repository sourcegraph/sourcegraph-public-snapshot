import { describe, expect, test } from '@jest/globals'

import { ExternalServiceKind } from '../../graphql-operations'

import type { ExternalServiceFieldsWithConfig } from './backend'
import { codeHostExternalServices, gitHubAppConfig, resolveExternalServiceCategory } from './externalServices'

describe('gitHubAppConfig', () => {
    test('get config', () => {
        const config = gitHubAppConfig('https://test.com', '1234', '5678')
        expect(config.defaultConfig).toEqual(`{
  "url": "https://test.com",
  "gitHubAppDetails": {
    "installationID": 5678,
    "appID": 1234,
    "baseURL": "https://test.com",
    "cloneAllRepositories": true
  },
  "authorization": {}
}`)
    })

    describe('resolveExternalServiceCategory', () => {
        test('parses config from JSONC if parsed config is not given', () => {
            const externalService = {
                config: '{"url": "https://github.com"}',
                kind: ExternalServiceKind.GITHUB,
            } as ExternalServiceFieldsWithConfig

            expect(resolveExternalServiceCategory(externalService)).toEqual(codeHostExternalServices.github)
        })

        test('returns GITHUB_DOTCOM for github.com', () => {
            const externalService = {
                parsedConfig: { url: 'https://github.com' },
                kind: ExternalServiceKind.GITHUB,
            } as ExternalServiceFieldsWithConfig

            expect(resolveExternalServiceCategory(externalService)).toEqual(codeHostExternalServices.github)
        })

        test('returns GHE if GITHUB kind and non-github.com URL', () => {
            const externalService = {
                parsedConfig: { url: 'https://gitlab.example.com' },
                kind: ExternalServiceKind.GITHUB,
            } as ExternalServiceFieldsWithConfig

            expect(resolveExternalServiceCategory(externalService)).toEqual(codeHostExternalServices.github)
        })

        test('returns GitLab dotcom if GITLAB kind and gitlab.com URL', () => {
            const externalService = {
                kind: ExternalServiceKind.GITLAB,
                parsedConfig: { url: 'https://gitlab.com' },
            } as ExternalServiceFieldsWithConfig

            expect(resolveExternalServiceCategory(externalService)).toEqual(codeHostExternalServices.gitlabcom)
        })

        test('returns GitLab category if GITLAB kind and non-gitlab.com URL', () => {
            const externalService = {
                kind: ExternalServiceKind.GITLAB,
                parsedConfig: { url: 'https://gitlab.example.com' },
            } as ExternalServiceFieldsWithConfig

            expect(resolveExternalServiceCategory(externalService)).toEqual(codeHostExternalServices.gitlabcom)
        })

        test('returns GitHub App category if config contains gitHubAppDetails', () => {
            const externalService = {
                kind: ExternalServiceKind.GITHUB,
                parsedConfig: {
                    url: 'https://github.com',
                    gitHubAppDetails: {
                        appID: 1234,
                        installationID: 5678,
                        baseURL: 'https://github.com',
                    },
                },
            } as ExternalServiceFieldsWithConfig
            const gitHubApp = { id: '1234', name: 'test-app' }

            expect(resolveExternalServiceCategory(externalService, gitHubApp)).toEqual({
                ...codeHostExternalServices.ghapp,
                additionalFormComponent: expect.anything(),
            })
        })
    })
})
