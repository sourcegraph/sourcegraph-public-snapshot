import * as React from "react";
import { Panel, RepositoryCard } from "sourcegraph/components";
import { whitespace } from "sourcegraph/components/utils";

export function RepositoryComponent(): JSX.Element {
	return <div>
		<Panel hoverLevel="low">
			<div style={{ padding: whitespace[5] }}>

				<RepositoryCard
					repo={{
						uri: "github.com/golang/go",
						description: "The Go Programming Language",
						language: "Go",
						private: false,
						pushedAt: "2016-11-10 19:17:31 +0000 UTC",
						fork: false,
						createdAt: "",
						__typename: "",
					} as GQL.IRemoteRepository} />

			</div>
			<hr />
			<code>
				<pre style={{
					whiteSpace: "pre-wrap",
					paddingLeft: whitespace[5],
					paddingRight: whitespace[5],
				}}>
					{`
	<RepositoryCard
		repo={{
			ID: 123,
			URI: "github.com/golang/go",
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
	</div>;
}
