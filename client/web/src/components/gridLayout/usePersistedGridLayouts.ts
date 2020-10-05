import { useCallback, useMemo } from 'react'
import { Layouts as ReactGridLayouts, ResponsiveProps } from 'react-grid-layout'
import { useLocalStorage } from '../../util/useLocalStorage'

type OnLayoutChange = NonNullable<ResponsiveProps['onLayoutChange']>

/**
 * A React hook that persists and restores grid layouts. When restoring, it keeps the positions of
 * the saved layouts but ignores minW, minH, and other things that can change.
 *
 * @example `const [layouts, onLayoutChange] = usePersistedGridLayouts('my-key', allDefaultLayouts)`
 */
export const usePersistedGridLayouts = (
    localStorageKey: string,
    allDefaultLayouts: ReactGridLayouts
): [ReactGridLayouts, OnLayoutChange] => {
    const [allSavedLayouts, setAllSavedLayouts] = useLocalStorage<ReactGridLayouts>(localStorageKey, allDefaultLayouts)

    const layouts = useMemo<ReactGridLayouts>(() => {
        for (const [breakpointName, defaultLayouts] of Object.entries(allDefaultLayouts)) {
            const savedLayouts = allSavedLayouts[breakpointName] || (allSavedLayouts[breakpointName] = defaultLayouts)
            for (const defaultLayout of defaultLayouts) {
                let savedLayout = savedLayouts.find(({ i }) => i === defaultLayout.i)
                if (!savedLayout) {
                    savedLayouts.push(defaultLayout)
                    savedLayout = defaultLayout
                }

                savedLayout.minW = defaultLayout.minW
                savedLayout.minH = defaultLayout.minH
            }
        }
        return allSavedLayouts
    }, [allSavedLayouts, allDefaultLayouts])

    const onLayoutChange = useCallback<OnLayoutChange>((_layout, allLayouts) => setAllSavedLayouts(allLayouts), [
        setAllSavedLayouts,
    ])

    return [layouts, onLayoutChange]
}
