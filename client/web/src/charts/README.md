## Sourcegraph Сharts

This package provides different visual
components mostly for rendering chart and data visualization.

At the moment this packages contains the following list of charts:

- **Series line chart** (with stacked and regular modes)
- **Pie chart**
- _Bar chart (is under development)_

As you will see in the storybook stories of this package, you can use these charts
directly from this package as the following example

```tsx
import { LineChart, LegendList, LegendItem } from './charts'

// Chart data
const SERIES = []
const DATA = {}

const Example = props => {
  return (
    <div>
      <ParentSize>
        {({ width, height }) => <LineChart width={width} height={height} data={DATA} series={SERIES} xAxisKey="x" />}
      </ParentSize>
      <LegendList>
        {SERIES.map(line => (
          <LegendItem key={line.dataKey.toString()} color={getLineColor(line)} name={line.name} />
        ))}
      </LegendList>
    </div>
  )
}
```

Or you can use view agnostic components where you can easily switch between
different types of charts as long as you can pick the right type of data
(series or categorical like, see the further section)

### Categorical vs Series

If we look at some charts like bar charts and pie charts, we
obviously will see that they have different visual representations, but
in terms of data for these charts, they are the same. Pie and bar chart are
both **categorical charts**. You can easily switch the pie chart with a bar
chart and change nothing in the data structure.

There is the same thing with around line and bar charts. Line charts represent
data changing through time, and bar char also can represent this data in the
same way. So they both are **series-like charts**. They both operate series
(lines or bars) as low-level data blocks.

```tsx
import { CategoricalChart, SeriesChart } from './charts'

const CategoricalPieChart = () => (
  <CategoricalChart
    // Here you could used other types (for example Bar chart)
    type={CategoricalBasedChartTypes.Pie}
    width={400}
    height={400}
    data={LANGUAGE_USAGE_DATA}
    getDatumName={getName}
    getDatumValue={getValue}
    getDatumColor={getColor}
    getDatumLink={getLink}
  />
)

const SeriesLineChart = () => (
  <SeriesChart
    type={SeriesBasedChartTypes.Line}
    data={DATA}
    series={SERIES}
    stacked={false}
    xAxisKey="x"
    width={400}
    height={400}
  />
)
```

### Direct imported charts vs Categorical/Series view-like charts

If you want to visualize some data and know exactly what chart you want to use for it,
you probably need to use directly imported charts (line, pie, bar charts). The main benefit
of using data grouped charts like Categorical and Series is that you can easily change the
view part (bar chart instead line chart) later. The chart view is just another setting for
a graph with Categorical and Series charts.
