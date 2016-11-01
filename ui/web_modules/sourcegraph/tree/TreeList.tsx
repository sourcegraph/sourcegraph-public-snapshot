import * as classNames from "classnames";
import * as React from "react";
import * as Relay from "react-relay";
import {Link} from "react-router";
import {urlToBlob} from "sourcegraph/blob/routes";
import {Base, Header, Heading, Panel} from "sourcegraph/components";
import {FileIcon, FolderIcon} from "sourcegraph/components/Icons";
import {typography} from "sourcegraph/components/utils";
import {whitespace} from "sourcegraph/components/utils/index";
import {urlToTree} from "sourcegraph/tree/routes";
import * as styles from "sourcegraph/tree/styles/Tree.css";
import "sourcegraph/tree/TreeBackend";

interface Props {
	repo: string;
	rev: string;
	path: string;
}

class TreeListComponent extends React.Component<Props & {root: GQL.IRoot}, {}> {
	render(): JSX.Element | null {
		const tree = this.props.root.repository.commit.tree;
		if (tree === null) {
			return <Header
					title="Not Found"
					subtitle="Directory not found." />;
		}

		let items: JSX.Element[] = [];
		if (this.props.path !== "/") {
			items.push(
				<Link className={classNames(styles.list_item, styles.parent_dir)}
					to={urlToTree(this.props.repo, this.props.rev, this.props.path.substr(0, this.props.path.lastIndexOf("/")))}
					key="$parent">
					<FolderIcon className={styles.icon} />
					..
				</Link>
			);
		}

		items = items.concat(tree.directories.map((dir) =>
			<Link className={classNames(styles.list_item)}
				to={urlToTree(this.props.repo, this.props.rev, this.props.path + "/" + dir.name)}
				key={dir.name}>
				<FolderIcon className={styles.icon} />
				{dir.name}
			</Link>
		));

		items = items.concat(tree.files.map((file) =>
			<Link className={classNames(styles.list_item)}
				to={urlToBlob(this.props.repo, this.props.rev, this.props.path + "/" + file.name)}
				key={file.name}>
				<FileIcon className={styles.icon} />
				{file.name}
			</Link>
		));

		const sx = Object.assign({},
			typography.size[5],
		);

		return <Panel style={sx}>
			<Base p={3} mb={3}>
				<Heading level={7} color="gray"
					style={{
						marginTop: whitespace[3],
						marginBottom: whitespace[3],
						marginLeft: whitespace[3],
					}}>Files</Heading>
				{items}
			</Base>
		</Panel>;
	}
}

const TreeListContainer = Relay.createContainer(TreeListComponent, {
	initialVariables: {
		repo: "",
		rev: "",
		path: "",
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				repository(uri: $repo) {
					commit(rev: $rev) {
						tree(path: $path) {
							directories {
								name
							}
							files {
								name
							}
						}
					}
				}
			}
		`,
	},
});

export const TreeList = (props: Props) => {
	return <Relay.RootContainer
		Component={TreeListContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: props,
		}}
	/>;
};
