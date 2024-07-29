import { describe, expect, test } from 'vitest'

import { BatchChangesCodeHostFields, ExternalServiceKind } from '../../../graphql-operations'

import { credentialForGitHubAppExists } from './github-apps-filter'

describe('credentialForGitHubAppExists', () => {
    describe('when connections are undefined', () => {
        test('it should yield false', () => {
            const result = credentialForGitHubAppExists(null, undefined)
            expect(result).toBe(false)
        })
    })

    describe('when there are no connections', () => {
        test('it should yield false', () => {
            const connections: BatchChangesCodeHostFields[] = []
            const result = credentialForGitHubAppExists(null, connections)
            expect(result).toBe(false)
        })
    })

    describe('when there is a connection for the targeted host', () => {

        describe('but it is not a github app', () => {
            test('it should yield false', () => {
                const appName = 'test';
                const connections: BatchChangesCodeHostFields[] = [
                    {
                        externalServiceKind: ExternalServiceKind.GITLAB,
                        externalServiceURL: 'https://gitlab.com',
                        requiresSSH: false,
                        requiresUsername: true,
                        supportsCommitSigning: true,
                        // The credential for gitlab technically can't have a github app, but we include this
                        // here so the test doesn't become a false positive.
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
                const result = credentialForGitHubAppExists(appName, connections)
                expect(result).toBe(false)
            })
        })

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
                const result = credentialForGitHubAppExists(null, connections)
                expect(result).toBe(false)
            })
        })

        describe('without any credential', () => {
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
                const result = credentialForGitHubAppExists('test', connections)
                expect(result).toBe(false)
            })
        })

        describe('with a credential', () => {
            describe('with a matching app name', () => {
                test('it should yield true', () => {
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
                    const result = credentialForGitHubAppExists(appName, connections)
                    expect(result).toBe(true)
                })
            })

            describe('without a matching app name', () => {
                test('it should yield false', () => {
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
                    const result = credentialForGitHubAppExists('differentAppName', connections)
                    expect(result).toBe(false)
                })
            })
        })
    })
})
