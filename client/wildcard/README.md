# Wildcard Component Library

## Overview

The Wildcard component library is a collection of design-approved reusable components that are suitable for use within the Sourcegraph codebase.

## Usage

This library is setup as a `yarn workspace` symlink.

You can import components from the library like so:

```javascript
import { PageSelector } from '@sourcegraph/wildcard'
```

## Folder Structure

- `src/`
  - `components/`
    Reusable React components
  - `hooks/`
    Reusable React hooks. Typically utilities or headless components

## Component Architecture

We want our components to be composable and reusable, but we don't want to 'reinvent the wheel' for everything we build. To support this, we aim to build our components with the following architecture as inspired by [React Spectrum](https://react-spectrum.adobe.com/).

<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="241" height="231" viewBox="-0.5 -0.5 241 231"><rect width="240" height="70" rx="10.5" ry="10.5" fill="#0fb6f2" stroke="#000" pointer-events="all"/><switch transform="translate(-.5 -.5)"><foreignObject style="overflow:visible;text-align:left" pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility"><div xmlns="http://www.w3.org/1999/xhtml" style="display:flex;align-items:unsafe center;justify-content:unsafe center;width:238px;height:1px;padding-top:35px;margin-left:1px"><div style="box-sizing:border-box;font-size:0;text-align:center"><div style="display:inline-block;font-size:16px;font-family:Helvetica;color:#fff;line-height:1.2;pointer-events:all;white-space:normal;word-wrap:normal">Component</div></div></div></foreignObject><text x="120" y="40" fill="#FFF" font-family="Helvetica" font-size="16" text-anchor="middle">Component</text></switch><rect y="80" width="240" height="70" rx="10.5" ry="10.5" fill="#b114f7" stroke="#000" pointer-events="all"/><switch transform="translate(-.5 -.5)"><foreignObject style="overflow:visible;text-align:left" pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility"><div xmlns="http://www.w3.org/1999/xhtml" style="display:flex;align-items:unsafe center;justify-content:unsafe center;width:238px;height:1px;padding-top:115px;margin-left:1px"><div style="box-sizing:border-box;font-size:0;text-align:center"><div style="display:inline-block;font-size:16px;font-family:Helvetica;color:#fff;line-height:1.2;pointer-events:all;white-space:normal;word-wrap:normal">Behaviour Hook</div></div></div></foreignObject><text x="120" y="120" fill="#FFF" font-family="Helvetica" font-size="16" text-anchor="middle">Behaviour Hook</text></switch><rect y="160" width="240" height="70" rx="10.5" ry="10.5" fill="#f86012" stroke="#000" pointer-events="all"/><switch transform="translate(-.5 -.5)"><foreignObject style="overflow:visible;text-align:left" pointer-events="none" width="100%" height="100%" requiredFeatures="http://www.w3.org/TR/SVG11/feature#Extensibility"><div xmlns="http://www.w3.org/1999/xhtml" style="display:flex;align-items:unsafe center;justify-content:unsafe center;width:238px;height:1px;padding-top:195px;margin-left:1px"><div style="box-sizing:border-box;font-size:0;text-align:center"><div style="display:inline-block;font-size:16px;font-family:Helvetica;color:#fff;line-height:1.2;pointer-events:all;white-space:normal;word-wrap:normal">State Hook</div></div></div></foreignObject><text x="120" y="200" fill="#FFF" font-family="Helvetica" font-size="16" text-anchor="middle">State Hook</text></switch><switch><a transform="translate(0 -5)" xlink:href="https://www.diagrams.net/doc/faq/svg-export-text-problems" target="_blank"></a></switch></svg>

### State hook

These hooks should act as a 'headless component', it should be entirely focused on state management and should function irrespective of any UI elements.

In many scenarios we may find that it makes sense to use third-party libraries for common patterns here.

### Behavior Hook

These hooks should capture key behavior and accessibility patterns that can be isolated from UI elements. For many simple components, this hook won't be necessary or requied. For larger components this becomes more meaningful.

For example, a `<Modal>` component might use `useToggleState` to handle displaying and hiding the modal, but we might have additional _behaviour_ like keyboard shortcuts and accessibility attributes that are relevant to this pattern. This is where something like `useModalBehavior` would make sense.

Like our state hooks, it is wise to use suitable third-party libraries for common patterns.

### Component

Now we have extracted our state and behaviour into seperate hooks, our UI component should just focus on displaying simple elements with specific styles.

**For most scenarios, it doesn't make sense to use third-party libraries here.** When compared to other applications, it will ultimately be our UI that will differ the most, not our UX.
