import React, { FunctionComponent } from 'react'
import { Timestamp } from '../../../components/time/Timestamp'
import { LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'
import { CodeIntelOptionalTimestamp } from '../shared/CodeIntelOptionalTimestamp'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeInteUploadOrIndexerRepository'

export interface CodeIntelIndexInfoProps {
    index: LsifIndexFields
    now?: () => Date
}

export const CodeIntelIndexInfo: FunctionComponent<CodeIntelIndexInfoProps> = ({ index, now }) => (
    <>
        <table className="table">
            <tbody>
                <tr>
                    <td>Repository</td>
                    <td>
                        <CodeIntelUploadOrIndexRepository node={index} />
                    </td>
                </tr>

                <tr>
                    <td>Commit</td>
                    <td>
                        <CodeIntelUploadOrIndexCommit node={index} abbreviated={false} />
                    </td>
                </tr>

                <tr>
                    <td>Queued</td>
                    <td>
                        <Timestamp date={index.queuedAt} now={now} noAbout={true} />
                    </td>
                </tr>

                <tr>
                    <td>Began processing</td>
                    <td>
                        <CodeIntelOptionalTimestamp
                            date={index.startedAt}
                            fallbackText="Index has not yet started."
                            now={now}
                        />
                    </td>
                </tr>

                <tr>
                    <td>
                        {index.state === LSIFIndexState.ERRORED && index.finishedAt ? 'Failed' : 'Finished'} processing
                    </td>
                    <td>
                        <CodeIntelOptionalTimestamp
                            date={index.finishedAt}
                            fallbackText="Index has not yet completed."
                            now={now}
                        />
                    </td>
                </tr>
            </tbody>
        </table>
    </>
)
