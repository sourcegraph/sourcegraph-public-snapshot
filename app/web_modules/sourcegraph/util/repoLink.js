import React from "react";
import * as router from "sourcegraph/util/router";
import Styles from "sourcegraph/components/styles/base.css";

// repoLink is swiped from app/repo.go.
function repoLink(repoURI, disableLink) {
	let collection = [],
		parts = repoURI.split("/");

	parts[0] = parts[0].toLowerCase();

	for (let i = 0; i < parts.length; i++) {
		if (i === 0 && parts.length > 1 && (parts[i] === "sourcegraph.com" || parts[i] === "github.com")) {
			continue;
		}
		if (i === parts.length - 1) {
			if (disableLink) {
				collection.push(<span className="name" key={`name${i}`} title={repoURI}>{parts[i]}</span>);
			} else {
				collection.push(<a className={Styles.link} key={`name${i}`} href={router.repo(repoURI)} title={repoURI}>{parts[i]}</a>);
			}
		} else {
			collection.push(<span className="part" key={`part${i}`}>{parts[i]}</span>);
		}
		if (i < parts.length - 1) {
			collection.push(<span className="sep" key={`sep${i}`}>/</span>);
		}
	}

	return <span className="repo-link">{collection}</span>;
}
export default repoLink;
