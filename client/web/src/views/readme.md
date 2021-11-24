## Shared view like components for building view gallery

This package provides standard low-level components for a view like
(currently insights) gallery pages.

- `<View />` - Rendering simple view card with different contents
- `<ViewGrid /` - Rendering gallery of cards on the page

## `<View />`

This component implements a visual view card block.
This component allows you to create a card with different types of content.
Check the `/view/View.story.tsx` for more details. Currently, this component is used
for rendering insights cards on the Code Insights dashboard page

**Empty view card**

```tsx
import * as View from './view'

/**
 * ┌─────────────────────┐
 * │Title text           │  View can have different view
 * ├─────────────────────┤  content (chart, error,
 * │░░░░░░░░░░░░░░░░░░░░░│  loading). See components
 * │░░░░░░░░░░░░░░░░░░░░░│  below
 * │░░░░░░░░░░░░░░░░░░░░░│
 * │░░░░View content░░░░░│
 * │░░░░░░░░░░░░░░░░░░░░░│
 * │░░░░░░░░░░░░░░░░░░░░░│
 * │░░░░░░░░░░░░░░░░░░░░░│
 * │░░░░░░░░░░░░░░░░░░░░░│
 * └─────────────────────┘
 */
function InsightView() {
  // Renders empty card component
  return <View.Root title="View title" />
}
```

**With chart content**

```tsx
import * as View from './view'

/**
 *
 * Renders view with chart content inside. Note that View.Content also support
 * markdown content.
 * ┌─────────────────────┐
 * │Title text           │
 * ├─────────────────────┤
 * │ ▲                   │
 * │ │             ●     │
 * │ │    ●       ●      │
 * │ │      ●    ●       │
 * │ │   ●    ●          │
 * │ │  ●                │
 * │ │                   │
 * │ └─────────────────▶ │
 * └─────────────────────┘
 */
function ChartView() {
  return (
    <View.Root>
      <View.Content view={VIEW_CHART_DATA} />
    </View.Root>
  )
}
```

**With loading content**

```tsx
import * as View from './View'

/**
 * ┌─────────────────────┐
 * │Title text           │
 * ├─────────────────────┤
 * │                     │
 * │                     │
 * │                     │
 * │   Loading content   │
 * │          ◕          │
 * │                     │
 * │                     │
 * │                     │
 * └─────────────────────┘
 */
function LoadingView() {
  return (
    <View.Root title="Title text">
      <View.LoadingContent />
    </View.Root>
  )
}
```

Usually, consumers mix these components above to build a rich view component with different states.
For example, this is one of view like components that Code Insights pages use for insights rendering.

```tsx
import * as View from './View'

function RichView() {
  const { data, loading } = useObservable(getBuiltInInsightData())

  return (
    <View.Root title="title text" className="extension-insight-card">
      {!data || loading || isDeleting ? (
        <View.LoadingContent
          text={isDeleting ? 'Deleting code insight' : 'Loading code insight'}
          subTitle={insight.id}
          icon={PuzzleIcon}
        />
      ) : isErrorLike(data.view) ? (
        <View.ErrorContent error={data.view} icon={PuzzleIcon} />
      ) : (
        data.view && <View.Content viewContent={data.view.content} viewID={insight.id} />
      )}
    </View.Root>
  )
}
```

## `<ViewGrid />`

This component renders a dynamic view gallery of `<View />` components. All views
within the grid are draggable and resizable. The current implementation uses [React Grid Layout](https://github.com/react-grid-layout/react-grid-layout)
library to implement a dynamic and responsive grid system.

```text
┌────────────┐ ┌────────────┐ ┌────────────┐
│▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │ │▪▪▪▪▪▪▪▪▪   │
│            │ │            │ │            │
│            │ │            │ │            │
│            │ │            │ │            │
│           ◿│ │           ◿│ │           ◿│
└────────────┘ └────────────┘ └────────────┘
┌───────────────────────────┐ ┌────────────┐
│■■■■■■■■■■■■■■■■■■■■■■■    │ │▪▪▪▪▪▪▪▪    │
│                           │ │            │
│                           │ │            │
│                           │ │            │
│                           │ │           ◿│
│                           │ └────────────┘
│                           │ ┌────────────┐
│                           │ │▪▪▪▪▪▪▪▪▪▪▪ │
│                           │ │            │
│                           │ │            │
│                           │ │            │
│                         ◿ │ │           ◿│
└───────────────────────────┘ └────────────┘
```

Simple example of usage `<ViewGrid />`

```tsx
import ViewGrid from './view-grid'
import * as View from '/view'

function ChartView(props) {
  return (
    <View.Root title={props.title}>
      <View.Content view={props.data} />
    </View.Root>
  )
}

function ViewGridExample() {
  return (
    <ViewGrid>
      <ChartView title="View #1" data={DATA_CHART} />
      <ChartView title="View #2/" data={DATA_CHART} />
      <ChartView title="View #3" data={DATA_CHART} />
    </ViewGrid>
  )
}
```
