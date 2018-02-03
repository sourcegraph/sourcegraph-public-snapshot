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
}

/**
 * A horizontal bar that displays tab titles, which the user can click to switch to the tab.
 *
 * @template T The type that includes all possible tab IDs (typically a union of string constants).
 */
export class TabBar<T extends string> extends React.PureComponent<TabBarProps<T>> {
    public render(): JSX.Element | null {
        return (
            <div className="tab-bar">
                {this.props.tabs.map((tab, i) => (
                    <button
                        key={i}
                        className={`btn btn-link btn-sm tab-bar__tab tab-bar__tab--${
                            this.props.activeTab === tab.id ? 'active' : 'inactive'
                        }`}
                        // tslint:disable-next-line:jsx-no-lambda
                        onClick={() => this.props.onSelect(tab.id)}
                    >
                        {tab.label}
                    </button>
                ))}
            </div>
        )
    }
}

interface TabsProps<T extends string> {
    /** All tabs. */
    tabs: Tab<T>[]

    /** A key unique to this UI element that is used for persisting the view state. */
    storageKey: string

    children: React.ReactElement<{ key: T }>[]

    className?: string
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
        return (
            <div className={`tabs ${this.props.className || ''}`}>
                <TabBar tabs={this.props.tabs} activeTab={this.state.activeTab} onSelect={this.onSelectTab} />
                {this.props.children.find(c => c.key === this.state.activeTab)}
            </div>
        )
    }

    private onSelectTab = (tab: T) =>
        this.setState({ activeTab: tab }, () => Tabs.saveToLocalStorage(this.props.storageKey, tab))
}
