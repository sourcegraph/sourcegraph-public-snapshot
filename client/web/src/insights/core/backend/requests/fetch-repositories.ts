import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql';

import { requestGraphQL } from '../../../../backend/graphql';
import { BulkSearchRepositories } from '../../../../graphql-operations';

const bulkSearchRepositoriesFragment = gql`
    fragment BulkSearchRepositories on RepositoryRedirect {
         ... on Repository {
            name
         }
        ... on Redirect {
            url
        }
    }
`

export function fetchRepositories(repositories: string[]): Observable<BulkSearchRepositories[]> {
    return requestGraphQL<Record<string, BulkSearchRepositories>>(`
        ${bulkSearchRepositoriesFragment}
        query BulkRepositoriesSearch(${repositories.map((repo, index) => `$query${index}: String!`).join(', ')}) {
            ${repositories
                .map((_repo, index) => `
                    repoSearch${index}: repositoryRedirect(name: $query${index}) {
                        ...BulkSearchRepositories
                    }
                `).join('\n')
            }
        }
    `,
        Object.fromEntries(repositories.map((query, index) => [`query${index}`, query]))
    ).pipe(
        map(dataOrThrowErrors),
        map(search =>
            Object.keys(search).reduce<BulkSearchRepositories[]>((result, key) => {
                const index = +key.slice('repoSearch'.length);

                result[index] = search[key];

                return result;
            }, [])
        )
    )
}
