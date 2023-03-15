import { QueryInfo } from '@sourcegraph/cody-common'

export interface IntentDetector {
    detect(text: string): Promise<QueryInfo>
}
