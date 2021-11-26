### git-combine

This command is used to synthesize a large monorepo from multiple
repositories. It does this by translating commits from every upstream into a
new commit containing the upstream contents as a sub directory.

See https://github.com/sgtest/megarepo

This is running in our production kubernetes cluster. https://github.com/sourcegraph/deploy-sourcegraph-dot-com/tree/release/configure/git-combine

#### Building

This is very rarely deployed, so it can be done manually with the build.sh script:

```shell
./build.sh 0.0.2
```

#### Configuration

This is the configuration of the repository running in production:

```shell
git remote add --no-tags -t main freeCodeCamp https://github.com/freeCodeCamp/freeCodeCamp.git
git remote add --no-tags -t main grafana https://github.com/grafana/grafana.git
git remote add --no-tags -t main rails https://github.com/rails/rails.git
git remote add --no-tags -t main sourcegraph https://github.com/sourcegraph/sourcegraph.git
git remote add --no-tags -t main vscode https://github.com/microsoft/vscode.git
git remote add --no-tags -t master azure-docs https://github.com/MicrosoftDocs/azure-docs.git
git remote add --no-tags -t master chromium https://github.com/chromium/chromium.git
git remote add --no-tags -t master elastic https://github.com/elastic/elasticsearch.git
git remote add --no-tags -t master flutter https://github.com/flutter/flutter.git
git remote add --no-tags -t master git https://github.com/git/git.git
git remote add --no-tags -t master gitlab https://gitlab.com/gitlab-org/gitlab.git
git remote add --no-tags -t master go https://github.com/golang/go.git
git remote add --no-tags -t master homebrew https://github.com/Homebrew/homebrew-core.git
git remote add --no-tags -t master kubernetes https://github.com/kubernetes/kubernetes.git
git remote add --no-tags -t master linux https://github.com/torvalds/linux.git
git remote add --no-tags -t master mongo https://github.com/mongodb/mongo.git
git remote add --no-tags -t master nixpkgs https://github.com/NixOS/nixpkgs.git
git remote add --no-tags -t master rust https://github.com/rust-lang/rust.git
git remote add --no-tags -t master tensorflow https://github.com/tensorflow/tensorflow.git

git remote add origin https://sourcegraph-bot:$ACCESS_TOKEN@github.com/sgtest/megarepo.git
```
