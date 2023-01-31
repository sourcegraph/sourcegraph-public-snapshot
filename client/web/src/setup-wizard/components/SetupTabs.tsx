import { createContext, FC, PropsWithChildren, useContext } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'

import { Button, Tab, TabList, TabPanel, Tabs, TabListProps, useTabsContext } from '@sourcegraph/wildcard'

import styles from './SetupTabs.module.scss'

interface SetupTabsContextData {
    onTabChange: (index: number) => void
}

const SetupTabsContext = createContext<SetupTabsContextData>({
    onTabChange: noop,
})

interface SetupTabsProps {
    activeTabIndex: number
    defaultActiveIndex: number
    onTabChange: (activeIndex: number) => void
}

/**
 * The root visual element for the setup Tabs UI wizard layout. Enforces
 * the right layout and internal state for completed, current and further setup steps
 */
export const SetupTabs: FC<PropsWithChildren<SetupTabsProps>> = props => {
    const { activeTabIndex, defaultActiveIndex, children, onTabChange } = props

    return (
        <SetupTabsContext.Provider value={{ onTabChange }}>
            <Tabs
                lazy={true}
                behavior="forceRender"
                size="large"
                index={activeTabIndex}
                defaultIndex={defaultActiveIndex}
                className={classNames(styles.tabs, 'mx-auto')}
                onChange={onTabChange}
            >
                {children}
            </Tabs>
        </SetupTabsContext.Provider>
    )
}

/** UI component to declare list of steps headers (tabs) UI */
export const SetupList: FC<PropsWithChildren<TabListProps>> = props => (
    <TabList {...props} className={classNames(styles.headerList, props.wrapperClassName)} />
)

interface SetupTabProps {
    index: number
}

export const SetupTab: FC<PropsWithChildren<SetupTabProps>> = props => {
    const { index, children } = props
    const { selectedIndex } = useTabsContext()

    return (
        <Tab
            disabled={index !== selectedIndex}
            className={classNames(styles.tab, { [styles.tabCompleted]: selectedIndex > index })}
        >
            {children}
        </Tab>
    )
}

export { TabPanel as SetupStep }

interface SetupStepActions {
    nextAvailable: boolean
    finish?: boolean
    onSkip?: () => void
    onComplete?: () => void
}

export const SetupStepActions: FC<SetupStepActions> = props => {
    const { nextAvailable, finish, onSkip, onComplete } = props

    const { selectedIndex } = useTabsContext()
    const { onTabChange } = useContext(SetupTabsContext)

    const isFirstStep = selectedIndex === 0

    return (
        <footer className={styles.actions}>
            {!finish && (
                <>
                    <Button variant="secondary" className={styles.actionsSkip} onClick={onSkip}>
                        Skip setup
                    </Button>
                    {!isFirstStep && (
                        <Button variant="secondary" outline={true} onClick={() => onTabChange(selectedIndex - 1)}>
                            Previous step
                        </Button>
                    )}
                    <Button variant="primary" disabled={!nextAvailable} onClick={() => onTabChange(selectedIndex + 1)}>
                        Next
                    </Button>
                </>
            )}

            {finish && (
                <Button variant="primary" className="ml-auto" onClick={onComplete}>
                    Go to search!
                </Button>
            )}
        </footer>
    )
}
