import classNames from 'classnames'
import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import React, { useMemo } from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Link } from '@sourcegraph/shared/src/components/Link'

import { Timestamp } from '../../components/time/Timestamp'
import { BatchChangeFields } from '../../graphql-operations'

import styles from './BatchSpec.module.scss'

/** Reports whether str is a valid JSON document. */
const isJSON = (string: string): boolean => {
    try {
        JSON.parse(string)
        return true
    } catch {
        return false
    }
}

export interface BatchSpecProps {
    originalInput: BatchChangeFields['currentSpec']['originalInput']
}

export const BatchSpec: React.FunctionComponent<BatchSpecProps> = ({ originalInput }) => {
    // JSON is valid YAML, so the input might be JSON. In that case, we'll highlight and indent it
    // as JSON. This is especially nice when the input is a "minified" (no extraneous whitespace)
    // JSON document that's difficult to read unless indented.
    const inputIsJSON = isJSON(originalInput)
    const input = useMemo(() => (inputIsJSON ? JSON.stringify(JSON.parse(originalInput), null, 2) : originalInput), [
        inputIsJSON,
        originalInput,
    ])

    return <CodeSnippet code={input} language={inputIsJSON ? 'json' : 'yaml'} className="mb-3" />
}

type BatchSpecMetaProps = Pick<BatchChangeFields, 'createdAt' | 'lastApplier' | 'lastAppliedAt' | 'name'> &
    BatchSpecProps

export const BatchSpecMeta: React.FunctionComponent<BatchSpecMetaProps> = ({
    createdAt,
    lastApplier,
    lastAppliedAt,
    name,
    originalInput,
}) => {
    const downloadUrl = useMemo(() => 'data:text/plain;charset=utf-8,' + encodeURIComponent(originalInput), [
        originalInput,
    ])

    return (
        <div className="d-flex flex-wrap justify-content-between align-items-baseline mb-2 test-batches-spec">
            <p className={classNames(styles.batchSpecTabHeaderCol, 'mb-2')}>
                {lastApplier ? <Link to={lastApplier.url}>{lastApplier.username}</Link> : 'A deleted user'}{' '}
                {createdAt === lastAppliedAt ? 'created' : 'updated'} this batch change{' '}
                <Timestamp date={lastAppliedAt} /> by applying the following batch spec:
            </p>
            <div className={styles.batchSpecTabHeaderCol}>
                <a
                    download={`${name}.batch.yaml`}
                    href={downloadUrl}
                    className="text-right btn btn-secondary text-nowrap"
                    data-tooltip={`Download ${name}.batch.yaml`}
                >
                    <FileDownloadIcon className="icon-inline" /> Download YAML
                </a>
            </div>
        </div>
    )
}
