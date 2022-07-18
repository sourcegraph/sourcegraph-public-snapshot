import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { FileDecoration } from 'sourcegraph'

import { fileDecorationColorForTheme } from '@sourcegraph/shared/src/api/extension/api/decorations'
import { Tooltip } from '@sourcegraph/wildcard'

import styles from './FileDecorator.module.scss'

interface FileDecoratorProps {
    /** Array of file decorations to render */
    fileDecorations?: FileDecoration[]

    /** File decorations may be styled differently depending on the theme */
    isLightTheme: boolean

    /**
     * File decorations may be styled differently depending on whether or not
     * the decorated file is active
     */
    isActive?: boolean

    className?: string
}

/**
 * Renders a list of file decorations from extensions
 */
export const FileDecorator: React.FunctionComponent<React.PropsWithChildren<FileDecoratorProps>> = ({
    fileDecorations,
    isLightTheme,
    className,
    isActive,
}) => {
    // Only need to check for number of decorations, other validation (like whether the decoration specifies at
    // least one of `text` or `percentage`) is done in the extension host
    if (!fileDecorations || fileDecorations.length === 0) {
        return null
    }

    return (
        <div
            className={classNames(
                'd-flex align-items-center text-nowrap pl-1 test-file-decoration-container',
                className
            )}
        >
            {fileDecorations.map(
                (fileDecoration, index) =>
                    (fileDecoration.meter || fileDecoration.after) && (
                        <div
                            className={classNames('d-flex align-items-center', styles.fileDecoration)}
                            key={fileDecoration.uri + String(index)}
                        >
                            {fileDecoration.after && (
                                <Tooltip content={fileDecoration.after.hoverMessage} placement="bottom">
                                    <small
                                        // eslint-disable-next-line react/forbid-dom-props
                                        style={{
                                            color: fileDecorationColorForTheme(
                                                fileDecoration.after,
                                                isLightTheme,
                                                isActive
                                            ),
                                        }}
                                        className={classNames(
                                            'text-monospace font-weight-normal test-file-decoration-text',
                                            styles.after,
                                            isActive && styles.afterActive
                                        )}
                                    >
                                        {fileDecoration.after.contentText}
                                    </small>
                                </Tooltip>
                            )}
                            {fileDecoration.meter && (
                                <Tooltip content={fileDecoration.meter.hoverMessage} placement="bottom">
                                    <div>
                                        <VisuallyHidden>{fileDecoration.meter.hoverMessage}</VisuallyHidden>
                                        <meter
                                            className={classNames('test-file-decoration-meter', styles.meter, {
                                                'ml-2': !!fileDecoration.after,
                                            })}
                                            min={fileDecoration.meter.min}
                                            low={fileDecoration.meter.low}
                                            high={fileDecoration.meter.high}
                                            max={fileDecoration.meter.max}
                                            optimum={fileDecoration.meter.optimum}
                                            value={fileDecoration.meter.value}
                                            aria-hidden={true}
                                        />
                                    </div>
                                </Tooltip>
                            )}
                        </div>
                    )
            )}
        </div>
    )
}
