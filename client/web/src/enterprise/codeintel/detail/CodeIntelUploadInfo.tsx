import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import React, { FunctionComponent } from 'react'
import { LSIFUploadState } from '../../../../../shared/src/graphql-operations'
import { Timestamp } from '../../../components/time/Timestamp'
import { LsifUploadFields } from '../../../graphql-operations'
import { CodeIntelOptionalTimestamp } from '../shared/CodeIntelOptionalTimestamp'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadRoot } from '../shared/CodeIntelUploadRoot'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeInteUploadOrIndexerRepository'

export interface CodeIntelUploadInfoProps {
    upload: LsifUploadFields
    now?: () => Date
}

export const CodeIntelUploadInfo: FunctionComponent<CodeIntelUploadInfoProps> = ({ upload, now }) => (
    <table className="table">
        <tbody>
            <tr>
                <td>Repository</td>
                <td>
                    <CodeIntelUploadOrIndexRepository node={upload} />
                </td>
            </tr>

            <tr>
                <td>Commit</td>
                <td>
                    <CodeIntelUploadOrIndexCommit node={upload} abbreviated={false} />
                </td>
            </tr>

            <tr>
                <td>Root</td>
                <td>
                    <CodeIntelUploadRoot node={upload} />
                </td>
            </tr>

            <tr>
                <td>Indexer</td>
                <td>
                    <CodeIntelUploadOrIndexIndexer node={upload} />
                </td>
            </tr>

            <tr>
                <td>
                    Is latest for repo{' '}
                    <small data-tooltip="This upload can answer queries for the tip of the default branch and are targets of cross-repository find reference operations.">
                        <InfoCircleOutlineIcon className="icon-inline cursor-pointer" />
                    </small>
                </td>
                <td>
                    {upload.finishedAt ? (
                        <span className="test-is-latest-for-repo">{upload.isLatestForRepo ? 'yes' : 'no'}</span>
                    ) : (
                        <span className="text-muted">Upload has not yet completed.</span>
                    )}
                </td>
            </tr>

            <tr>
                <td>Uploaded</td>
                <td>
                    <Timestamp date={upload.uploadedAt} now={now} noAbout={true} />
                </td>
            </tr>

            <tr>
                <td>Began processing</td>
                <td>
                    <CodeIntelOptionalTimestamp
                        date={upload.startedAt}
                        fallbackText="Upload has not yet started."
                        now={now}
                    />
                </td>
            </tr>

            <tr>
                <td>
                    {upload.state === LSIFUploadState.ERRORED && upload.finishedAt ? 'Failed' : 'Finished'} processing
                </td>
                <td>
                    <CodeIntelOptionalTimestamp
                        date={upload.finishedAt}
                        fallbackText="Upload has not yet completed."
                        now={now}
                    />
                </td>
            </tr>
        </tbody>
    </table>
)
