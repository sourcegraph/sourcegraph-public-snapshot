import React from 'react'
import classNames from 'classnames'
import { FileDecoration } from 'sourcegraph'
import { fileDecorationColorForTheme } from '../../../shared/src/api/client/services/decoration'

interface FileDecoratorProps {
    /** Array of file decorations to render */
    fileDecorations?: FileDecoration[]

    /** File decorations may be styled differently depending on the theme */
    isLightTheme: boolean

    /**
     * File decorations may be styled differently depending on whether or not
     * the decorated file is selected
     */
    isSelected?: boolean

    className?: string
}

/**
 * Renders a list of file decorations from extensions
 */
export const FileDecorator: React.FunctionComponent<FileDecoratorProps> = ({
    fileDecorations,
    isLightTheme,
    className,
    isSelected,
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
                            className="file-decoration d-flex align-items-center"
                            key={fileDecoration.path + String(index)}
                        >
                            {fileDecoration.after && (
                                <small
                                    // eslint-disable-next-line react/forbid-dom-props
                                    style={{
                                        color: fileDecorationColorForTheme(
                                            fileDecoration.after,
                                            isLightTheme,
                                            isSelected
                                        ),
                                    }}
                                    data-tooltip={fileDecoration.after.hoverMessage}
                                    data-placement="bottom"
                                    className="file-decoration__after text-monospace font-weight-normal test-file-decoration-text"
                                >
                                    {fileDecoration.after.value}
                                </small>
                            )}
                            {fileDecoration.meter && (
                                <meter
                                    className={classNames('file-decoration__meter test-file-decoration-meter', {
                                        'ml-2': !!fileDecoration.after,
                                    })}
                                    min={fileDecoration.meter.min}
                                    low={fileDecoration.meter.low}
                                    high={fileDecoration.meter.high}
                                    max={fileDecoration.meter.max}
                                    optimum={fileDecoration.meter.optimum}
                                    value={fileDecoration.meter.value}
                                    data-tooltip={fileDecoration.meter.hoverMessage}
                                    data-placement="bottom"
                                />
                            )}
                        </div>
                    )
            )}
        </div>
    )
}
