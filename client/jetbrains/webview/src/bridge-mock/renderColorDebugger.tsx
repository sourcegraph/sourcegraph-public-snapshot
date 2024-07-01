/* eslint-disable react/forbid-dom-props */
import { createRoot } from 'react-dom/client'

import { Code, H2 } from '@sourcegraph/wildcard'

import { dark } from './theme-snapshots/dark'
import { light } from './theme-snapshots/light'

export const renderColorDebugger = (): void => {
    document.body.innerHTML = "<div id='color-debug'></div>"

    const root = createRoot(document.querySelector('#color-debug')!)

    root.render(<ColorDebugger />)
}

const ColorDebugger = (): JSX.Element => (
    <div
        style={{
            display: 'flex',
            alignItems: 'start',
            justifyContent: 'center',
            margin: 20,
        }}
    >
        <div>
            <H2>IntelliJ Theme</H2>
            <ColorPalette dark={dark.intelliJTheme} light={light.intelliJTheme} />
        </div>
    </div>
)

const ColorPalette = ({
    dark,
    light,
}: {
    dark: { [key: string]: string }
    light: { [key: string]: string }
}): JSX.Element => {
    const colors: Map<string, { dark: null | string; light: null | string }> = new Map()

    for (const [key, value] of Object.entries(dark)) {
        colors.set(key, { dark: value, light: null })
    }
    for (const [key, value] of Object.entries(light)) {
        if (colors.has(key)) {
            colors.get(key)!.light = value
        } else {
            colors.set(key, { dark: null, light: value })
        }
    }

    return (
        <>
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
                    <Color hexValue={dark} name={key} />
                    <Color hexValue={light} name={key} />
                </div>
            ))}
        </>
    )
}

const Color = ({ hexValue, name }: { hexValue: string | null; name: string }): JSX.Element => (
    <div
        style={{
            backgroundColor: hexValue ?? 'transparent',
            color: hexValue ? getColorForBackgroundColor(hexValue) : 'black',
            width: '50%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
        }}
    >
        {name} {hexValue !== null ? <Code>({hexValue})</Code> : null}
    </div>
)

function getColorForBackgroundColor(backgroundColor: string): string {
    const color = backgroundColor.slice(1) // strip     #
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
