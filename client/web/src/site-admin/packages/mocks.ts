import { ApolloError } from '@apollo/client'

import { PackageRepoMatchesResult, PackageRepoMatchesVariables } from '../../graphql-operations'

interface usePackageRepoMatchesMockResult {
    data?: PackageRepoMatchesResult
    loading: boolean
    error?: ApolloError
}

interface usePackageRepoMatchesMockVariables {
    variables: PackageRepoMatchesVariables
    skip?: boolean
}

export const usePackageRepoMatchesMock = ({
    variables,
    skip,
}: usePackageRepoMatchesMockVariables): usePackageRepoMatchesMockResult => {
    // TODO: Uncomment
    // const { data, loading, error } = useQuery<PackageRepoMatchesResult, PackageRepoMatchesVariables>(
    //     packageRepoMatchesQuery,
    //     {
    //         variables: {
    //             scheme: PackageRepoReferenceKind.NPMPACKAGES,
    //             filter: {
    //                 nameMatcher: {
    //                     packageGlob: namePattern,
    //                 },
    //             },
    //             first: 15,
    //         },
    //         skip,
    //     }
    // )

    // return { data, loading, error }

    if (skip) {
        return {
            loading: false,
        }
    }

    // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const getMockVersion = () => ({
        id: Math.floor(Math.random() * 1000).toString(),
        version: `1.0.${Math.floor(Math.random() * 1000)}`,
    })

    // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const getMockNode = () => ({
        id: Math.floor(Math.random() * 1000).toString(),
        name: `@types/jest-${Math.floor(Math.random() * 1000)}`,
        repository: {
            id: Math.floor(Math.random() * 1000).toString(),
            url: '/npm/types/jest',
            name: 'npm/types/jest',
            mirrorInfo: {
                byteSize: Math.floor(Math.random() * 100000).toString(),
            },
        },
        versions: Array.from({ length: 15 }, getMockVersion),
    })

    return {
        loading: false,
        data: {
            packageReposMatches: {
                nodes: Array.from({ length: 15 }, getMockNode),
                totalCount: 200,
            },
        },
    }
}
