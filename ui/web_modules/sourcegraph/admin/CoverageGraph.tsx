// tslint:disable

import * as React from "react";

import {LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ReferenceLine} from "recharts";

type Props = {
	data: {
		Day: string,
		Refs: number,
		Defs: number,
	}[],
	target?: number,
}

export default class CoverageGraph extends React.Component<Props, any> {
	render(): JSX.Element | null {
		return (
			<LineChart width={600} height={300} data={this.props.data}>
				<XAxis dataKey="Day"/>
				<YAxis />
				<CartesianGrid strokeDasharray="3 3"/>
				<Tooltip />
				<Legend />
				{this.props.target && <ReferenceLine y={this.props.target} stroke="red"/>}
				<Line type="monotone" dataKey="Refs" stroke="#8884d8" />
				<Line type="monotone" dataKey="Defs" stroke="#82ca9d" />
			</LineChart>
		);
	}
}
