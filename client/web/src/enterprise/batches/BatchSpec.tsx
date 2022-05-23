import React, { useMemo } from 'react'

import { kebabCase } from 'lodash'
import FileDownloadIcon from 'mdi-react/FileDownloadIcon'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, Button, Icon } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'
import { BatchChangeFields } from '../../graphql-operations'

import { MonacoBatchSpecEditor } from './batch-spec/edit/editor/MonacoBatchSpecEditor'

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
    name: string
    originalInput: BatchChangeFields['currentSpec']['originalInput']
    className?: string
}

export const BatchSpec: React.FunctionComponent<React.PropsWithChildren<BatchSpecProps>> = ({
    originalInput,
    isLightTheme,
    className,
    name,
}) => {
    // JSON is valid YAML, so the input might be JSON. In that case, we'll highlight and indent it
    // as JSON. This is especially nice when the input is a "minified" (no extraneous whitespace)
    // JSON document that's difficult to read unless indented.
    const inputIsJSON = isJSON(originalInput)
    const input = useMemo(() => (inputIsJSON ? JSON.stringify(JSON.parse(originalInput), null, 2) : originalInput), [
        inputIsJSON,
        originalInput,
    ])

    return (
        <MonacoBatchSpecEditor
            batchChangeName={name}
            isLightTheme={isLightTheme}
            value={input}
            readOnly={true}
            className={className}
        />
    )
}

interface BatchSpecDownloadLinkProps extends BatchSpecProps, Pick<BatchChangeFields, 'name'> {
    className?: string
    asButton: boolean
}

export const BatchSpecDownloadLink: React.FunctionComponent<
    React.PropsWithChildren<BatchSpecDownloadLinkProps>
> = React.memo(function BatchSpecDownloadLink({ children, className, name, originalInput, asButton }) {
    return asButton ? (
        <Button
            variant="primary"
            as="a"
            download={getFileName(name)}
            href={'data:text/plain;charset=utf-8,' + encodeURIComponent(originalInput)}
            target="_blank"
            rel="noopener noreferrer"
            className={className}
            data-tooltip={`Download ${getFileName(name)}`}
        >
            {children}
        </Button>
    ) : (
        <Link
            download={getFileName(name)}
            to={'data:text/plain;charset=utf-8,' + encodeURIComponent(originalInput)}
            className={className}
            data-tooltip={`Download ${getFileName(name)}`}
        >
            {children}
        </Link>
    )
})

// TODO: Consider merging this component with BatchSpecDownloadLink
export const BatchSpecDownloadButton: React.FunctionComponent<
    React.PropsWithChildren<BatchSpecProps & Pick<BatchChangeFields, 'name'>>
> = React.memo(function BatchSpecDownloadButton(props) {
    return (
        <Button
            className="text-right text-nowrap"
            {...props}
            variant="secondary"
            outline={true}
            as={BatchSpecDownloadLink}
            asButton={false}
        >
            <Icon as={FileDownloadIcon} /> Download YAML
        </Button>
    )
})

type BatchSpecMetaProps = Pick<BatchChangeFields, 'createdAt' | 'lastApplier' | 'lastAppliedAt'>

export const BatchSpecMeta: React.FunctionComponent<React.PropsWithChildren<BatchSpecMetaProps>> = ({
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
