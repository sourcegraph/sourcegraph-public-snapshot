import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss';
import { PieExample } from './PieChart';
import { LineChartExample } from './Threshold';

const { add } = storiesOf('web/VisxCharts', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        {/* Chart will always fill the container, so we need to give the container an explicit size. */}
        <div style={{ width: '32rem', height: '16rem' }}>{story()}</div>
    </>
))

add('Pie chart', () => (
    <PieExample width={550} height={350}/>
))

add('Line chart', () => (
    <LineChartExample width={550} height={350} />
))
