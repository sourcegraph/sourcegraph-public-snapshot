/* eslint-disable react/forbid-dom-props */
import ReactDOM from 'react-dom'

import { dark } from './theme-snapshots/dark'
import { light } from './theme-snapshots/light'

export const renderColorDebugger = (): void => {
    document.body.innerHTML = "<div id='color-debug'></div>"
    ReactDOM.render(<ColorDebugger />, document.querySelector('#color-debug'))
}

const ColorDebugger = (): JSX.Element => {
    const colors: Map<string, { dark: null | string; light: null | string }> = new Map()

    for (const [key, value] of Object.entries(dark)) {
        colors.set(key, { dark: value, light: null })
    }
    for (const [key, value] of Object.entries(light)) {
        if (colors.has(key)) {
            // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            colors.get(key)!.light = value
        } else {
            colors.set(key, { dark: null, light: value })
        }
    }

    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: 20,
            }}
        >
            <div>
                {Array.from(colors).map(([key, { dark, light }]) => (
                    <div
                        key={key}
                        style={{
                            display: 'flex',
                            flexDirection: 'row',
                            height: 30,
                            width: 800,
                            fontFamily: '-apple-system, BlinkMacSystemFont, sans-serif',
                            fontSize: 14,
                            fontWeight: 'bold',
                        }}
                    >
                        <div
                            style={{
                                backgroundColor: dark ?? 'transparent',
                                color: dark ? getColorForBackgroundColor(dark) : 'black',
                                width: '50%',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                            }}
                        >
                            {key}
                        </div>
                        <div
                            style={{
                                backgroundColor: light ?? 'transparent',
                                color: light ? getColorForBackgroundColor(light) : 'black',
                                width: '50%',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                            }}
                        >
                            {key}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    )
}

function getColorForBackgroundColor(backgroundColor: string): string {
    const color = backgroundColor.slice(1) // strip #
    const rgb = parseInt(color, 16) // convert rrggbb to decimal
    const red = (rgb >> 16) & 0xff // extract red
    const green = (rgb >> 8) & 0xff // extract green
    const blue = (rgb >> 0) & 0xff // extract blue

    const luma = 0.2126 * red + 0.7152 * green + 0.0722 * blue // per ITU-R BT.709

    if (luma < 40) {
        return '#ffffff'
    }
    return '#000000'
}
