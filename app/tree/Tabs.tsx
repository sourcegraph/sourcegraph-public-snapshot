import * as React from 'react'

/** Describes a tab. */
export interface Tab<T extends string> {
    id: T
    label: string
}

interface TabBarProps<T extends string> {
    /** All tabs. */
    tabs: Tab<T>[]

    /** The currently active tab. */
    activeTab: T

    /** Called when the user selects a different tab. */
    onSelect: (tab: T) => void

    /** A fragment to render at the end of the tab bar. */
    endFragment?: React.ReactFragment

    tabClassName?: string
}

/**
 * A horizontal bar that displays tab titles, which the user can click to switch to the tab.
 *
 * @template T The type that includes all possible tab IDs (typically a union of string constants).
 */
class TabBar<T extends string> extends React.PureComponent<TabBarProps<T>> {
    public render(): JSX.Element | null {
        return (
            <div className="tab-bar">
                {this.props.tabs.map((tab, i) => (
                    <button
                        key={i}
                        className={`tab-btn tab-bar__tab ${!this.props.endFragment &&
                            'tab-bar__tab--flex-grow'} tab-bar__tab--${
                            this.props.activeTab === tab.id ? 'active' : 'inactive'
                        } ${this.props.tabClassName || ''}`}
                        // tslint:disable-next-line:jsx-no-lambda
                        onClick={() => this.props.onSelect(tab.id)}
                    >
                        {tab.label}
                    </button>
                ))}
                {this.props.endFragment}
            </div>
        )
    }
}

/**
 * An element to pass to Tab's tabBarEndFragment prop to fill all width between the tabs (on the left) and the
 * other tabBarEndFragment elements (on the right).
 */
export const Spacer: () => JSX.Element | any = () => <span className="tab-bar__spacer" />

interface TabsProps<T extends string> {
    /** All tabs. */
    tabs: Tab<T>[]

    /** A key unique to this UI element that is used for persisting the view state. */
    storageKey: string

    /**
     * A fragment to display at the end of the tab bar. If specified, the tabs will not flex grow to fill the
     * width.
     */
    tabBarEndFragment?: React.ReactFragment

    children: undefined | React.ReactElement<{ key: T }> | (undefined | React.ReactElement<{ key: T }>)[]

    id?: string
    className?: string
    tabClassName?: string
}

interface TabState<T extends string> {
    /** The currently active tab. */
    activeTab: T
}

/**
 * A tabbed UI component, with a tab bar for switching between tabs and a content view that renders the active
 * tab's contents.
 */
export class Tabs<T extends string> extends React.PureComponent<TabsProps<T>, TabState<T>> {
    /**
     * The class name to use for other elements injected via tabBarEndFragment that should have a bottom border.
     */
    public static tabBorderClassName = 'tab-bar__end-fragment-other-element'

    constructor(props: TabsProps<T>) {
        super(props)

        this.state = {
            activeTab: Tabs.readFromLocalStorage(this.props.storageKey, this.props.tabs),
        }
    }

    private static readFromLocalStorage<T extends string>(storageKey: string, tabs: Tab<T>[]): T {
        const lastTabID = localStorage.getItem(storageKey)
        if (lastTabID !== null && tabs.find(tab => tab.id === lastTabID)) {
            return lastTabID as T
        }
        return tabs[0].id // default
    }

    private static saveToLocalStorage<T extends string>(storageKey: string, lastTab: T): void {
        localStorage.setItem(storageKey, lastTab)
    }

    public render(): JSX.Element | null {
        let children: React.ReactElement<{ key: T }>[] | undefined
        if (Array.isArray(this.props.children)) {
            children = this.props.children as React.ReactElement<{ key: T }>[]
        } else if (this.props.children) {
            children = [this.props.children as React.ReactElement<{ key: T }>]
        }

        return (
            <div id={this.props.id} className={`tabs ${this.props.className || ''}`}>
                <TabBar
                    tabs={this.props.tabs}
                    activeTab={this.state.activeTab}
                    onSelect={this.onSelectTab}
                    endFragment={this.props.tabBarEndFragment}
                    tabClassName={this.props.tabClassName}
                />
                {children && children.find(c => c && c.key === this.state.activeTab)}
            </div>
        )
    }

    private onSelectTab = (tab: T) =>
        this.setState({ activeTab: tab }, () => Tabs.saveToLocalStorage(this.props.storageKey, tab))
}
