import * as React from "react";
import {Code, Heading, Panel, RepositoryCard, Table} from "sourcegraph/components";
import {whitespace} from "sourcegraph/components/utils";

export function RepositoryComponent(): JSX.Element {
	return <div style={{marginBottom: whitespace[4], marginTop: whitespace[4]}}>
		<Heading level={3}>Repository Card</Heading>

		<Panel hoverLevel="low">
			<div style={{padding: whitespace[4]}}>

				<RepositoryCard
					repo={{
						uri: "https://sourcegraph.com/github.com/golang/go",
						owner: "golang",
						name: "go",
						description: "The Go Programming Language",
						language: "Go",
						private: false,
						pushedAt: "2016-11-10 19:17:31 +0000 UTC",
						fork: false,
						createdAt: "",
						mirror: false,
						__typename: "",
						httpCloneURL: "https://sourcegraph.com/github.com/golang/go",
						vcsSyncedAt: "",
					}}
					contributors={[
						{
							AvatarURL: "https://avatars.githubusercontent.com/u/8691941?v=3",
							UID: "8691941",
							Login: "jamescuddy",
						},
						{
							AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
							UID: "8691941",
							Login: "chexee",
						},
						{
							AvatarURL: "https://avatars.githubusercontent.com/u/8691941?v=3",
							UID: "8691941",
							Login: "jamescuddy",
						},
						{
							AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
							UID: "8691941",
							Login: "chexee",
						},
						{
							AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
							UID: "8691941",
							Login: "chexee",
						},
						{
							AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
							UID: "8691941",
							Login: "chexee",
						},
						{
							AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
							UID: "8691941",
							Login: "chexee",
						},
					]} />

			</div>
			<hr />
			<code>
				<pre style={{
					whiteSpace: "pre-wrap",
					paddingLeft: whitespace[4],
					paddingRight: whitespace[4],
				}}>
{`
	<RepositoryCard
		repo={{
			ID: 123,
			URI: "https://sourcegraph.com/github.com/golang/go",
			Owner: "golang",
			Name: "go",
			Description: "The Go Programming Language",
			DefaultBranch: "master",
			Language: "Go",
			Private: false,
		}}
		contributors={[
			{
				AvatarURL: "https://avatars.githubusercontent.com/u/8691941?v=3",
				UID: "8691941",
				Login: "jamescuddy",
			},
			{
				AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
				UID: "8691941",
				Login: "chexee",
			},
			{
				AvatarURL: "https://avatars.githubusercontent.com/u/8691941?v=3",
				UID: "8691941",
				Login: "jamescuddy",
			},
			{
				AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
				UID: "8691941",
				Login: "chexee",
			},
			{
				AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
				UID: "8691941",
				Login: "chexee",
			},
			{
				AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
				UID: "8691941",
				Login: "chexee",
			},
			{
				AvatarURL: "https://avatars.githubusercontent.com/u/285836?v=3",
				UID: "8691941",
				Login: "chexee",
			},
		]} />
`
}
				</pre>
			</code>
		</Panel>
		<Heading level={6} style={{marginTop: whitespace[3], marginBottom: whitespace[2]}}>
			RepositoryCard Properties
		</Heading>
		<Panel hoverLevel="low" style={{padding: whitespace[4]}}>
			<Table style={{width: "100%"}}>
				<thead>
					<tr>
						<td>Prop</td>
						<td>Default value</td>
						<td>Values</td>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td><Code>repo</Code></td>
						<td><Code>undefined</Code></td>
						<td>
							The repository object. Required.
						</td>
					</tr>
					<tr>
						<td><Code>contributors</Code></td>
						<td><Code>undefined</Code></td>
						<td>
							An array of user objects. Required.
						</td>
					</tr>
				</tbody>
			</Table>
		</Panel>
	</div>;
}
