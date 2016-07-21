import React from "react";
import TimeAgo from "sourcegraph/util/TimeAgo";
import {Avatar} from "sourcegraph/components";
import CSSModules from "react-css-modules";
import styles from "./styles/Commit.css";

function showBothSigs(a, b) {
	return a && b && (a.Name !== b.Name || a.Email !== b.Email);
}

function username(email) {
	if (!email) return "(unknown)";
	const i = email.indexOf("@");
	if (i === -1) return email;
	return `${email.slice(0, i)}@…`;
}

function sigName(sig) {
	if (!sig) return null;
	if (sig.Name) {
		return (
			<span styleName="sig-name">
				{sig.Name ? <span styleName="name">{sig.Name}&nbsp;</span> : null}
				<span styleName="name-secondary">({username(sig.Email)})</span>
			</span>
		);
	}
	return (
		<span styleName="sig-name">
			<span styleName="name">{username(sig.Email)}</span>
		</span>
	);
}

function Commit({repo, commit, full}) {
	const parts = commit.Message ? commit.Message.split("\n") : null;
	let title = parts ? parts[0] : "(no commit message)";
	let rest = parts ? parts.slice(1).join("\n") : null;
	rest = rest.trim();

	const maxTitleSize = 120;
	if (title.length > maxTitleSize) {
		rest = `…${title.slice(maxTitleSize)}\n${rest}`;
		title = `${title.slice(0, maxTitleSize)}…`;
	}

	return (
		<div styleName="container">
			<div styleName="main">
				<span styleName="title">{title}</span>
				<div styleName="meta">
					<Avatar className={styles.avatar} img={commit.AuthorPerson ? commit.AuthorPerson.AvatarURL : ""} size="small" />
					<div styleName="signature">
						<span styleName="sig">{sigName(commit.Author)} authored <TimeAgo time={commit.Author.Date} /></span><wbr/>
						{commit.Committer && showBothSigs(commit.Author, commit.Committer) ? <span styleName="sig">{sigName(commit.Committer)} committed <TimeAgo time={commit.Committer.Date} /></span> : null}
					</div>
					<div styleName="commit-id">
						<code styleName="sha">{commit.ID.substring(0, 8)}</code>
					</div>
				</div>
			</div>
			{full && rest && <div styleName="rest">{rest}</div>}
		</div>
	);
}
Commit.propTypes = {
	commit: React.PropTypes.object.isRequired,
	repo: React.PropTypes.string.isRequired,

	// full is whether to show the full commit message (beyond the first line).
	full: React.PropTypes.bool.isRequired,
};

export default CSSModules(Commit, styles);
