import React from 'react'
import classNames from 'classnames'
import { FileDecoration } from 'sourcegraph'
import { fileDecorationColorForTheme } from '../../../shared/src/api/client/services/decoration'

/**
 *
 * NOTE: Don't use `FileDecorator` like a component (pass to React.createElement).
 * Instead, call it directly. It's meant to return ReactNodes, which
 * includes null.
 */
export function FileDecorator({
    fileDecorations,
    isLightTheme,
    isDirectory,
    isSelected,
}: {
    fileDecorations?: FileDecoration[]
    isLightTheme: boolean
    isDirectory?: boolean
    isSelected?: boolean
}): React.ReactNode {
    // Only need to check for number of decorations, other validation (like whether the decoration specifies at
    // least one of `text` or `percentage`) is done in the extension host
    if (!fileDecorations || fileDecorations.length === 0) {
        return null
    }

    return (
        <div
            className={classNames('d-flex align-items-center text-nowrap test-file-decoration-container', {
                'mr-3': isDirectory,
            })}
        >
            {fileDecorations.map(
                (fileDecoration, index) =>
                    (fileDecoration.meter || fileDecoration.after) && (
                        <div
                            className="file-decoration d-flex align-items-center"
                            key={fileDecoration.path + String(index)}
                        >
                            {fileDecoration.after && (
                                // link or span?
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
