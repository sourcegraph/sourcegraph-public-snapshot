import * as autobind from "autobind-decorator";
import * as bluebird from "bluebird";
import * as update from "immutability-helper";
import * as _ from "lodash";
import * as React from "react";

interface User {
	login: string;
}

interface Label {
	name: string;
}

interface Issue {
	id: string;
	number: number;
	labels: Label[];
	html_url: string;
	title: string;
	assignee: User | null;
	assignees: User[] | null;
}

interface PullRequest {
	number: number;
	labels: Label[];
	assignee: User | null;
	assignees: User[] | null;
}

function getAssignee(issue: Issue): string | null {
	if (issue.assignees && issue.assignees.length > 0) {
		return issue.assignees.map((user) => user.login).join(", ");
	}
	return issue.assignee ? issue.assignee.login : null;
}

function getProjectOrComponent(issue: Issue): string | null {
	if (issue.labels) {
		for (const label of issue.labels) {
			if (label.name.startsWith("Project: ") || label.name.startsWith("Component: ")) {
				return label.name;
			}
		}
	}
	return null;
}

function getPriority(issue: Issue): string | null {
	if (issue.labels) {
		for (const label of issue.labels) {
			if (label.name === "P1" || label.name === "P2" || label.name === "P3" || label.name === "P4" || label.name === "P5") {
				return label.name;
			}
		}
	}
	return null;
}

function getType(issue: Issue, type: string): string | null {
	if (issue.labels) {
		for (const label of issue.labels) {
			if (label.name.startsWith("Type: ")) {
				return label.name;
			}
		}
	}
	return null;
}

function isType(issue: Issue, type: string): boolean {
	if (issue.labels) {
		for (const label of issue.labels) {
			if (label.name === type) {
				return true;
			}
		}
	}
	return false;
}

function isBug(issue: Issue): boolean {
	return isType(issue, "Type: Bug");
}

function isFeature(issue: Issue): boolean {
	return isType(issue, "Type: Feature");
}

function isEnhancement(issue: Issue): boolean {
	return isType(issue, "Type: Enhancement");
}

function isDebt(issue: Issue): boolean {
	return isType(issue, "Type: Tech Debt");
}

function isChore(issue: Issue): boolean {
	return isType(issue, "Type: Chore");
}

const pullRequests: { [num: number]: PullRequest } = {};
function isPullRequest(issue: Issue): boolean {
	return Boolean(pullRequests[issue.number]);
}

interface Resp {
	Projects: { [name: string]: Project };
	OtherIssues: Issue[];
	PullRequests: PullRequest[];
}

interface Project {
	Name: string;
	Columns: { [name: string]: Column };
}

interface Column {
	ID: number;
	Name: string;
	Issues: Issue[];
}

interface State {
	loaded: boolean;
	showOverview: boolean;
	filterInput: string;
	fetching: boolean;
	groupBy: string[];
	filters: { [name: string]: string };
	types: { [name: string]: boolean };
	triageBugs: Issue[];
	backlogBugs: Issue[];
	todoBugs: Issue[];
	inProgressBugs: Issue[];
	verifyBugs: Issue[];
	closedBugs: Issue[];
	projects: { [name: string]: Project };
}

const allGroups = ["assignee", "board", "priority"];

@autobind
export class ProjectsOverview extends React.Component<{}, State> {

	constructor(props: any) {
		super(props);

		const initFilters = {};
		allGroups.forEach((group) => initFilters[group] = "");
		const initTypes = {
			bug: true,
			debt: true,
			feature: true,
			enhancement: true,
			chore: true,
			pulls: true,
		};

		this.state = {
			loaded: false,
			fetching: true,
			showOverview: false,
			groupBy: [],
			filters: initFilters,
			types: initTypes,
			filterInput: "",
			triageBugs: [],
			backlogBugs: [],
			todoBugs: [],
			inProgressBugs: [],
			verifyBugs: [],
			closedBugs: [],
			projects: {},
		};

		this.fetchData(false);
	}

	fetchData(setLoading: boolean): void {
		if (setLoading) {
			this.setState({ fetching: true } as State);
		}
		fetch("https://issues.sgdev.org/extension")
			.then((resp) => resp.json().then((data) => {
				const d = data as Resp;

				if (d.PullRequests) {
					d.PullRequests.forEach((pr) => pullRequests[pr.number] = pr);
				}

				let triage: Issue[] = [];
				let backlog: Issue[] = [];
				let todo: Issue[] = [];
				let inProgress: Issue[] = [];
				let verify: Issue[] = [];
				let closed: Issue[] = [];

				for (const projectName of Object.keys(d.Projects)) {
					const project = d.Projects[projectName];
					triage = triage.concat(project.Columns["Needs Triage"].Issues);
					backlog = backlog.concat(project.Columns["Backlog"].Issues);
					todo = todo.concat(project.Columns["TODO"].Issues);
					inProgress = inProgress.concat(project.Columns["In Progress"].Issues);
					verify = verify.concat(project.Columns["Verify"].Issues);
					closed = closed.concat(project.Columns["Closed"].Issues);
				}

				this.setState({
					fetching: false,
					loaded: true,
					triageBugs: _.compact(triage),
					backlogBugs: _.compact(backlog),
					todoBugs: _.compact(todo),
					inProgressBugs: _.compact(inProgress),
					verifyBugs: _.compact(verify),
					closedBugs: _.compact(closed),
					projects: d.Projects,
				} as State);
			}));
	}

	groupElements(issues: Issue[], group: (issue: Issue) => string): any[] {
		const groups = _.groupBy(issues, group);
		return _.sortBy(Object.keys(groups), (k) => k.toLowerCase()).map((g, i) => {
			return <div key={i} style={{ marginBottom: "10px" }}>
				<div style={{ fontWeight: "bold", marginBottom: "1px" }}>({groups[g].length}) {g}</div>
				{groups[g].map((bug, j) => {
					return <div key={j} style={{ display: "flex", alignItems: "center", borderBottom: "solid rgba(0, 0, 0, .2) 1px", paddingTop: "1px", paddingBottom: "1px" }}>
						<a style={{ flex: .8 }} href={bug.html_url}>#{bug.number}</a>
						{this.getTypeLabel(bug)}
						<span style={{ flex: 1.5, textAlign: "center", fontSize: "9px", color: getAssignee(bug) ? "black" : "red" }}>{getAssignee(bug) || "(no assignee)"}</span>
						<span style={{ flex: 10 }}> {bug.title}</span>
					</div>;
				})}
			</div>;
		});
	}

	groupByFunction(issue: Issue): string {
		return this.state.groupBy.map((group) => {
			switch (group) {
				case "assignee":
					return getAssignee(issue) || "(no assignee)";
				case "board":
					return getProjectOrComponent(issue) || "(no project)";
				case "priority":
					return getPriority(issue) || "(no priority)";
			}
			return "";
		}).join(" : ");
	}

	filterFunction(issue: Issue): boolean {
		// First remove excluded types.
		if (isBug(issue) && !this.state.types["bug"]) {
			return false;
		}
		if (isDebt(issue) && !this.state.types["debt"]) {
			return false;
		}
		if (isChore(issue) && !this.state.types["chore"]) {
			return false;
		}
		if (isFeature(issue) && !this.state.types["feature"]) {
			return false;
		}
		if (isEnhancement(issue) && !this.state.types["enhancement"]) {
			return false;
		}
		if (isPullRequest(issue) && !this.state.types["pulls"]) {
			return false;
		}

		for (const group of this.state.groupBy) {
			switch (group) {
				case "assignee":
					const assignee = getAssignee(issue);
					const assigneeFilter = this.state.filters["assignee"];
					if (!assignee && assigneeFilter) {
						return false;
					}
					if (assignee && assigneeFilter && assignee.toLowerCase().indexOf(assigneeFilter.toLowerCase()) === -1) {
						return false;
					}
				case "board":
					const board = getProjectOrComponent(issue);
					const boardFilter = this.state.filters["board"];
					if (!board && boardFilter) {
						return false;
					}
					if (board && boardFilter && board.toLowerCase().indexOf(boardFilter.toLowerCase()) === -1) {
						return false;
					}
				case "priority":
					const priority = getPriority(issue);
					const priorityFilter = this.state.filters["priority"];
					if (!priority && priorityFilter) {
						return false;
					}
					if (priority && priorityFilter && priority.toLowerCase().indexOf(priorityFilter.toLowerCase()) === -1) {
						return false;
					}
			}
		}

		return true;
	}

	typeToggles(): JSX.Element[] {
		return Object.keys(this.state.types).map((type, i) => {
			return <div key={i} style={{ marginLeft: "10px", marginRight: "10px", cursor: "pointer", display: "inline-block" }}
				onClick={() => this.setState(update(this.state, { types: { $merge: { [type]: !this.state.types[type] } } }))}>
				<span style={{ fontWeight: this.state.types[type] ? "bold" : "normal" }}>{type}</span>
			</div>;
		});
	}

	groupByToggles(): JSX.Element[] {
		return _.difference(allGroups, this.state.groupBy).map((val, i) => {
			return <div key={i} style={{ marginLeft: "10px", marginRight: "10px", cursor: "pointer", display: "inline-block" }}
				onClick={() => this.setState(Object.assign({}, this.state, { groupBy: this.state.groupBy.concat(val) }))}>
				<span>{val}</span>
			</div>;
		});
	}

	groupByTogglesActivated(): JSX.Element[] {
		return this.state.groupBy.map((val, i) => {
			return <div key={i} style={{ marginLeft: "10px", marginRight: "10px", cursor: "pointer", display: "inline-block" }}>
				<span onClick={() => this.setState(Object.assign({}, this.state, { groupBy: update(this.state.groupBy, { $splice: [[i, 1]] }) }))}>{val}</span>
				<input style={{ width: "80px", marginLeft: "5px" }} type="text" onChange={(e) => this.handleFilterChange(e, val)} value={this.state.filters[val]} />
			</div>;
		});
	}

	handleFilterChange(e: any, group: string): void {
		if (e && e.target) {
			const value = e.target.value;
			this.setState({ filters: update(this.state.filters, { $merge: { [group]: value } }) } as State);
		}
	}

	getTypeLabel(issue: Issue): JSX.Element | null {
		const style = { flex: .1, textAlign: "center", fontSize: "16px" };
		if (isBug(issue)) {
			return <span style={style}>🐞</span>;
		}
		if (isDebt(issue)) {
			return <span style={style}>⛓</span>;
		}
		if (isChore(issue)) {
			return <span style={style}>⚙</span>;
		}
		if (isEnhancement(issue)) {
			return <span style={style}>💵</span>;
		}
		if (isFeature(issue)) {
			return <span style={style}>💰</span>;
		}
		if (isPullRequest(issue)) {
			return <span style={style}>🚢</span>;
		}

		return <span style={style}>⚠</span>;
	}

	getGroup(issues: Issue[], columnName: string): JSX.Element {
		issues = issues.filter(this.filterFunction);
		return <div style={{ display: "flex", marginTop: "10px" }}>
			<div style={{ flex: 1, textDecoration: "underline", fontSize: "16px", fontWeight: "bold" }}>({issues.length}) {columnName}</div>
			<div style={{ flex: 5 }}>
				{this.state.groupBy.length !== 0 && this.groupElements(issues, (issue) => this.groupByFunction(issue))}
				{this.state.groupBy.length === 0 && issues.map((bug, i) =>
					<div key={i} style={{ display: "flex", alignItems: "center", borderBottom: "solid rgba(0, 0, 0, .2) 1px", paddingTop: "1px", paddingBottom: "1px" }}>
						<a style={{ flex: .8 }} href={bug.html_url}>#{bug.number}</a>
						{this.getTypeLabel(bug)}
						<span style={{ flex: 1.5, textAlign: "center", fontSize: "9px", color: getAssignee(bug) ? "black" : "red" }}>{getAssignee(bug) || "(no assignee)"}</span>
						<span style={{ flex: 10 }}> {bug.title}</span>
					</div>)}
			</div>
		</div>;
	}

	render(): JSX.Element | null {
		return <div style={{ marginBottom: "10px", display: "inline-block", width: "100%" }}>
			<strong style={{ cursor: this.state.loaded ? "pointer" : "" }} onClick={() => {
				if (!this.state.loaded) {
					return;
				}
				this.setState({ showOverview: !this.state.showOverview } as State);
			} }>{this.state.loaded ? `${this.state.showOverview ? "Hide" : "Show"} Issue Overview` : "Loading issue stats"}</strong>
			<span style={{ fontSize: "11px", marginLeft: "15px", cursor: !this.state.fetching ? "pointer" : "normal", fontWeight: "bold", float: "right" }}
				onClick={() => this.fetchData(true)}>
				{this.state.fetching ? "..." : "Refresh Data"}
			</span>
			{this.state.showOverview && <div style={{ marginTop: "10px", display: "flex" }}>
				<div style={{ fontWeight: "bold", fontSize: "11px" }}>TYPES:</div>
				{this.typeToggles()}
			</div>}
			{this.state.showOverview && <div style={{ marginTop: "10px", display: "flex" }}>
				<div style={{ flex: 1 }}>
					<div style={{ fontWeight: "bold", fontSize: "11px" }}>ADD GROUPS:</div>
					{this.groupByToggles()}
				</div>
				{this.state.groupBy.length > 0 && <div style={{ flex: 3 }}>
					<div style={{ fontWeight: "bold", fontSize: "11px" }}>ACTIVE GROUPS:</div>
					{this.groupByTogglesActivated()}
				</div>}
			</div>}
			{this.state.showOverview && <div style={{ marginBottom: "10px" }}>
				{this.getGroup(this.state.inProgressBugs, "In Progress")}
				{this.getGroup(this.state.verifyBugs, "Verify")}
				{this.getGroup(this.state.todoBugs, "TODO")}
				{this.getGroup(this.state.triageBugs, "Needs Triage")}
				{this.getGroup(this.state.backlogBugs, "Backlog")}
			</div>}
		</div>;
	}
}
