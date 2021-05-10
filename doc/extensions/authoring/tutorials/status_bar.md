# Status bar item tutorial

![Status bar](img/status-bar.png)

Extensions can display information in the status bar adjacent to code editors. The status bar UI elements contributed by extensions are called status bar items.

> Note: This feature was introduced in Sourcegraph version 3.26. 
> Extensions should check if:

>   - `sourcegraph.app.createStatusBarItemType`

>   - `setStatusBarItem` method on CodeEditors

> are defined to prevent errors on older versions of Sourcegraph. 


## `StatusBarItemType`

You'll need to create a `StatusBarItemType` to identify your status bar item: 

```ts
const statusBarItemType = sourcegraph.app.createStatusBarItemType()
```

Each editor can display one status bar item per type. If your extension needs to display multiple status bar items, create multiple `StatusBarItemType`.

## Getting a reference to an editor

Status bar items are properties of a code editor, so you'll need a reference to a code editor: 

```ts
import * as sourcegraph from 'sourcegraph'

// Check if the Sourcegraph instance supports status bar items (minimum Sourcegraph version 3.26)
const statusBarItemType = sourcegraph.app.createStatusBarItemType && sourcegraph.app.createStatusBarItemType()

export function activate(context: sourcegraph.ExtensionContext): void {
    sourcegraph.app.activeWindow?.activeViewComponentChanges.subscribe(viewComponent => {
        // Check if the view component is a code editor and supports status bar items.
        if (viewComponent?.type === 'CodeEditor' && 'setStatusBarItem' in viewComponent) {
            viewComponent.setStatusBarItem(statusBarItemType, { text: 'my status bar item' })
        }
    })
}
```

## Tooltips

Status bar items can display tooltips on hover:

```ts
editor.setStatusBarItem(statusBarItemType, { text: 'my status bar item', tooltip: 'my tooltip' })
```

## Commands

You can run [commands](https://unpkg.com/sourcegraph@25.2.0/dist/docs/modules/_sourcegraph_.commands.html) in response to click events. Here's how you can use the [built-in](http://docs.sourcegraph.com/extensions/authoring/builtin_commands#builtin-sourcegraph-extension-commands) `open` command to open Sourcegraph documentation:

```ts
editor.setStatusBarItem(statusBarItemType, { text: 'docs', command: { id: 'open', args: ['https://docs.sourcegraph.com'] } })
```

## Clearing the status bar item

Hide your status bar item by setting `text` to a falsy value:

```ts
editor.setStatusBarItem(statusBarItemType, { text: '' })
```
