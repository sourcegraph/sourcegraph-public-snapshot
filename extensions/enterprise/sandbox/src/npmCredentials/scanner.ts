import { flatten } from 'lodash'
import { from } from 'rxjs'
import { toArray } from 'rxjs/operators'
import { TextSearchResult } from 'sourcegraph'
import { memoizedFindTextInFiles } from '../util'
import { NPMCredentialsCampaignContext } from './providers'

export const scanForCredentials = async ({ filters }: NPMCredentialsCampaignContext): Promise<TextSearchResult[]> =>
    flatten(
        await from(
            memoizedFindTextInFiles(
                {
                    pattern: `/(^|:)_(auth|authToken|password)\\s*=\\s*[^\\s]+/ ${filters || ''}`,
                    type: 'regexp',
                },
                {
                    files: {
                        includes: ['(^|/)\\.npmrc$'],
                        type: 'regexp',
                    },
                    maxResults: 25, // TODO!(sqs): increase
                }
            )
        )
            .pipe(toArray())
            .toPromise()
    )
