## Сharts

This package provides different visual
components primarily for rendering charts and data visualization.

At the moment, this package contains the following list of charts:

- **Series line chart** (with stacked and regular modes)
- **Pie chart**
- **Categorical Bar chart**

As you can see in the storybook stories of this package, you can use these charts
directly from this package as the following example

```tsx
import { LineChart, LegendList, LegendItem } from './charts'

const SERIES = [
  {
    id: 'series_001',
    data: [
      { x: new Date(10, 10, 2020), value: 94 },
      { x: new Date(11, 10, 2020), value: 134 },
      { x: new Date(12, 10, 2020), value: 134 },
      { x: new Date(13, 10, 2020), value: 123 },
    ],
    name: 'Series 1',
    color: 'var(--blue)',
    getXValue: datum => datum.x,
    getYValue: datum => datum.value,
  },
]

const Example = props => {
  return (
    <div>
      <ParentSize>{({ width, height }) => <LineChart width={width} height={height} series={SERIES} />}</ParentSize>
      <LegendList>
        {SERIES.map(line => (
          <LegendItem key={line.dataKey.toString()} color={getLineColor(line)} name={line.name} />
        ))}
      </LegendList>
    </div>
  )
}
```
