import React from "react";
import {Link} from "react-router";
import {urlToRepo} from "sourcegraph/repo/routes";
import style from "sourcegraph/components/styles/breadcrumb.css";

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
				collection.push(<Link className={style.link} key={`name${i}`} to={urlToRepo(repoURI)} title={repoURI}>{parts[i]}</Link>);
			}
		} else {
			collection.push(<span className="part" key={`part${i}`}>{parts[i]}</span>);
		}
		if (i < parts.length - 1) {
			collection.push(<span className={style.sep} key={`sep${i}`}>/</span>);
		}
	}

	return <span className="repo-link">{collection}</span>;
}
export default repoLink;
