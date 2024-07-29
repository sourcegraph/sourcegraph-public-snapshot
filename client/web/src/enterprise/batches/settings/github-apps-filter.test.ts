import { describe, expect, test } from 'vitest'

import { BatchChangesCodeHostFields, ExternalServiceKind } from '../../../graphql-operations'

import { noCredentialForGitHubAppExists } from './github-apps-filter'

describe('github-apps-filter', () => {
    describe('when connections are undefined', () => {
        test('it should yield false', () => {
            const result = noCredentialForGitHubAppExists(null, undefined)
            expect(result).toBe(false)
        })
    })

    describe('when there are no connections', () => {
        test('it should yield false', () => {
            const connections: BatchChangesCodeHostFields[] = []
            const result = noCredentialForGitHubAppExists(null, connections)
            expect(result).toBe(false)
        })
    })

    describe('when there is a connection for the targeted host', () => {
        describe('without a target app name', () => {
            test('it should yield false', () => {
                const connections: BatchChangesCodeHostFields[] = [
                    {
                        externalServiceKind: ExternalServiceKind.GITHUB,
                        externalServiceURL: 'https://github.com',
                        requiresSSH: false,
                        requiresUsername: true,
                        supportsCommitSigning: true,
                        credential: null,
                        commitSigningConfiguration: null,
                    },
                ]
                const result = noCredentialForGitHubAppExists(null, connections)
                expect(result).toBe(false)
            })
        })

        describe('without any credential', () => {
            test('it should yield true', () => {
                const connections: BatchChangesCodeHostFields[] = [
                    {
                        externalServiceKind: ExternalServiceKind.GITHUB,
                        externalServiceURL: 'https://github.com',
                        requiresSSH: false,
                        requiresUsername: true,
                        supportsCommitSigning: true,
                        credential: null,
                        commitSigningConfiguration: null,
                    },
                ]
                const result = noCredentialForGitHubAppExists('test', connections)
                expect(result).toBe(true)
            })
        })

        describe('with a credential', () => {
            describe('with a matching app name', () => {
                test('it should yield false', () => {
                    const appName = 'test'
                    const connections: BatchChangesCodeHostFields[] = [
                        {
                            externalServiceKind: ExternalServiceKind.GITHUB,
                            externalServiceURL: 'https://github.com',
                            requiresSSH: false,
                            requiresUsername: true,
                            supportsCommitSigning: true,
                            credential: {
                                id: '1',
                                sshPublicKey: null,
                                isSiteCredential: false,
                                gitHubApp: {
                                    id: '1',
                                    appID: 1,
                                    name: appName,
                                    appURL: 'https://github.com',
                                    baseURL: 'https://github.com',
                                    logo: 'https://github.com',
                                },
                            },
                            commitSigningConfiguration: null,
                        },
                    ]
                    const result = noCredentialForGitHubAppExists(appName, connections)
                    expect(result).toBe(false)
                })
            })

            describe('without a matching app name', () => {
                test('it should yield true', () => {
                    const connections: BatchChangesCodeHostFields[] = [
                        {
                            externalServiceKind: ExternalServiceKind.GITHUB,
                            externalServiceURL: 'https://github.com',
                            requiresSSH: false,
                            requiresUsername: true,
                            supportsCommitSigning: true,
                            credential: {
                                id: '1',
                                sshPublicKey: null,
                                isSiteCredential: false,
                                gitHubApp: {
                                    id: '1',
                                    appID: 1,
                                    name: 'test',
                                    appURL: 'https://github.com',
                                    baseURL: 'https://github.com',
                                    logo: 'https://github.com',
                                },
                            },
                            commitSigningConfiguration: null,
                        },
                    ]
                    const result = noCredentialForGitHubAppExists('differentAppName', connections)
                    expect(result).toBe(true)
                })
            })
        })
    })
})
