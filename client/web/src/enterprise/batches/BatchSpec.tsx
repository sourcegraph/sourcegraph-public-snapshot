import { kebabCase } from 'lodash'
import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import React, { useMemo } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { Timestamp } from '../../components/time/Timestamp'
import { BatchChangeFields } from '../../graphql-operations'

import { MonacoBatchSpecEditor } from './create/editor/MonacoBatchSpecEditor'

/** Reports whether `string` is a valid JSON document. */
const isJSON = (string: string): boolean => {
    try {
        JSON.parse(string)
        return true
    } catch {
        return false
    }
}

export const getFileName = (name: string): string => `${kebabCase(name)}.batch.yaml`

export interface BatchSpecProps extends ThemeProps {
    originalInput: BatchChangeFields['currentSpec']['originalInput']
    className?: string
}

export const BatchSpec: React.FunctionComponent<BatchSpecProps> = ({ originalInput, isLightTheme, className }) => {
    // JSON is valid YAML, so the input might be JSON. In that case, we'll highlight and indent it
    // as JSON. This is especially nice when the input is a "minified" (no extraneous whitespace)
    // JSON document that's difficult to read unless indented.
    const inputIsJSON = isJSON(originalInput)
    const input = useMemo(() => (inputIsJSON ? JSON.stringify(JSON.parse(originalInput), null, 2) : originalInput), [
        inputIsJSON,
        originalInput,
    ])

    return <MonacoBatchSpecEditor isLightTheme={isLightTheme} value={input} readOnly={true} className={className} />
}

interface BatchSpecDownloadLinkProps extends BatchSpecProps, Pick<BatchChangeFields, 'name'> {
    className?: string
}

export const BatchSpecDownloadLink: React.FunctionComponent<BatchSpecDownloadLinkProps> = React.memo(
    function BatchSpecDownloadLink({ children, className, name, originalInput }) {
        return (
            <a
                download={getFileName(name)}
                href={'data:text/plain;charset=utf-8,' + encodeURIComponent(originalInput)}
                className={className}
                data-tooltip={`Download ${getFileName(name)}`}
            >
                {children}
            </a>
        )
    }
)

export const BatchSpecDownloadButton: React.FunctionComponent<
    BatchSpecProps & Pick<BatchChangeFields, 'name'>
> = React.memo(function BatchSpecDownloadButton(props) {
    return (
        <BatchSpecDownloadLink className="text-right btn btn-outline-secondary text-nowrap" {...props}>
            <FileDownloadIcon className="icon-inline" /> Download YAML
        </BatchSpecDownloadLink>
    )
})

type BatchSpecMetaProps = Pick<BatchChangeFields, 'createdAt' | 'lastApplier' | 'lastAppliedAt'>

export const BatchSpecMeta: React.FunctionComponent<BatchSpecMetaProps> = ({
    createdAt,
    lastApplier,
    lastAppliedAt,
}) => (
    <p className="mb-2">
        {lastApplier ? <Link to={lastApplier.url}>{lastApplier.username}</Link> : 'A deleted user'}{' '}
        {createdAt === lastAppliedAt ? 'created' : 'updated'} this batch change{' '}
        <Timestamp date={lastAppliedAt ?? createdAt} /> by applying the following batch spec:
    </p>
)
