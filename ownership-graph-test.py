import json
import os
import argparse

import networkx as nx
from networkx.algorithms import bipartite

def main(args):
    repo = args.repo

    file_authors = list([])

    if not repo:
        for root, dirs, _ in os.walk("/Users/erik/.sourcegraph/repos_1"):
            for name in dirs:
                if not os.path.isdir(os.path.join(root, name, ".git")):
                    continue
                repo_dir = os.path.join(root, name, ".git")

                repo_name = os.path.join(root, name)[len("/Users/erik/.sourcegraph/repos_1/"):]

                if not os.path.isfile(os.path.join(repo_dir, "author_data.json")):
                    continue

                with open(os.path.join(repo_dir, "author_data.json"), encoding="utf-8") as f:
                    print("loaded", f.name)
                    repo_file_authors = json.load(f)["data"]

                    # Add repo name prefix.
                    for file in repo_file_authors:
                        file["file"] = repo_name + "/" + file["file"]

                    if args.filter_files:
                        for f in [ file for file in repo_file_authors if args.filter_files in file["file"] ]:
                            file_authors.append(f)
                    else:
                        for f in repo_file_authors:
                            file_authors.append(f)

        for root, dirs, _ in os.walk("/Users/erik/.sourcegraph/repos_2"):
            for name in dirs:
                if not os.path.isdir(os.path.join(root, name, ".git")):
                    continue
                repo_dir = os.path.join(root, name, ".git")

                repo_name = os.path.join(root, name)[len("/Users/erik/.sourcegraph/repos_2/"):]

                if not os.path.isfile(os.path.join(repo_dir, "author_data.json")):
                    continue

                with open(os.path.join(repo_dir, "author_data.json"), encoding="utf-8") as f:
                    print("loaded", f.name)
                    repo_file_authors = json.load(f)["data"]

                    # Add repo name prefix.
                    for file in repo_file_authors:
                        file["file"] = repo_name + "/" + file["file"]

                    if args.filter_files:
                        for f in [ file for file in repo_file_authors if args.filter_files in file["file"] ]:
                            file_authors.append(f)
                    else:
                        for f in repo_file_authors:
                            file_authors.append(f)

    else:
        repo_dir = "/Users/erik/.sourcegraph/repos_1/"+repo+"/.git"
        if not os.path.isdir(repo_dir):
            repo_dir = "/Users/erik/.sourcegraph/repos_2/"+repo+"/.git"
            if not os.path.isdir(repo_dir):
                raise "repo not found"

        with open(os.path.join(repo_dir, "author_data.json"), encoding="utf-8") as f:
            file_authors = json.load(f)["data"]
            if args.filter_files:
                file_authors = [
                    file for file in file_authors if args.filter_files in file["file"]
                ]

    file_nodes = [file["file"] for file in file_authors]

    author_nodes = list(
        set([author["author"] for file in file_authors for author in file["authors"]])
    )

    B = nx.Graph()

    B.add_nodes_from(file_nodes, bipartite=0)
    B.add_nodes_from(author_nodes, bipartite=1)

    B.add_weighted_edges_from(
        [
            (file["file"], author["author"], author["changes"])
            for file in file_authors
            for author in file["authors"]
        ]
    )

    def my_weight(G, u, v, weight="weight"):
        w = 0
        for nbr in set(G[u]) & set(G[v]):
            w += G[u][nbr].get(weight, 1) + G[v][nbr].get(weight, 1)
        return w

    G_weighted = bipartite.generic_weighted_projected_graph(
        B, author_nodes, weight_function=my_weight
    )

    G_collab = bipartite.collaboration_weighted_projected_graph(B, author_nodes)

    if args.author:
        print("Collaboration")
        for neigh in sorted(
            G_collab[args.author].items(),
            key=lambda edge: edge[1]["weight"],
            reverse=True,
        )[:10]:
            print(neigh)

        print("Weighted by common changes")
        for neigh in sorted(
            G_weighted[args.author].items(),
            key=lambda edge: edge[1]["weight"],
            reverse=True,
        )[:10]:
            print(neigh)
    else:
        print("Degree centrality")
        degree_centrality = nx.degree_centrality(G_collab)
        for node, score in sorted(
            degree_centrality.items(), key=lambda x: x[1], reverse=True
        )[:10]:
            print(node, score)

        print("Pagerank")
        pagerank = nx.pagerank(G_collab)
        for node, score in sorted(pagerank.items(), key=lambda x: x[1], reverse=True)[
            :10
        ]:
            print(node, score)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", dest="repo", required=False)
    parser.add_argument("--author", dest="author", required=False)
    parser.add_argument("--filter_files", dest="filter_files", required=False)
    args = parser.parse_args()

    main(args)
