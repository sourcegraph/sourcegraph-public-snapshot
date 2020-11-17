import React, { FunctionComponent } from 'react'
import { Timestamp } from '../../../components/time/Timestamp'
import { LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'
import { CodeIntelOptionalTimestamp } from '../shared/CodeIntelOptionalTimestamp'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexRoot } from '../shared/CodeIntelUploadOrIndexRoot'
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
                    <td>Root</td>
                    <td>
                        <CodeIntelUploadOrIndexRoot node={index} />
                    </td>
                </tr>

                <tr>
                    <td>Indexer</td>
                    <td>
                        <CodeIntelUploadOrIndexIndexer node={index} />
                    </td>
                </tr>

                <tr>
                    <td>Indexer args</td>
                    <td>
                        <code>{index.indexerArgs.join(' ')}</code>
                    </td>
                </tr>

                <tr>
                    <td>Outfile</td>
                    <td>{index.outfile}</td>
                </tr>

                <tr>
                    <td>Docker steps</td>
                    <td>
                        {index.dockerSteps.length === 0 ? (
                            <span>No docker steps configured.</span>
                        ) : (
                            <table className="table">
                                <thead>
                                    <tr>
                                        <th>Root</th>
                                        <th>Image</th>
                                        <th>Commands</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {index.dockerSteps.map(step => (
                                        <tr key={`${step.root}${step.image}${step.commands.join(' ')}`}>
                                            <td>{step.root}</td>
                                            <td>{step.image}</td>
                                            <td>
                                                <code>{step.commands.join(' ')}</code>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        )}
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

                <tr>
                    <td>Log contents</td>
                    <td>
                        <pre>{index.logContents}</pre>
                    </td>
                </tr>
            </tbody>
        </table>
    </>
)
