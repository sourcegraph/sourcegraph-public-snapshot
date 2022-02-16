#!/bin/bash

branches=(
  ef/experiment-non-sequential-migrations
  ef/simplify-embed
  ef/refactor-sg-migration
  ef/inline-runner-visitor
  ef/refactor-cliutil
  ef/make-migration-store-transactional
  ef/remove-create-index-concurrently-down
  ef/migration-operations
  ef/improve-test-try-lock
  ef/add-index-status
  ef/new-definitions
  ef/simplify-sqlf-comparions
  ef/hierarchical-migrations
  ef/refactor-store
  ef/with-migration-log
  ef/migration-store-versions
  ef/migration-definition-parents
  ef/migration-definition-utils
  ef/humanize-migration-names
  ef/cleanup-metadata
  ef/remove-field
  ef/migration-operation-type
  ef/idempotent-unlock
  ef/squash-remove-all
  ef/fix-squash
  ef/squash-ignore-migration-logs
  ef/expose-migration-operations
  ef/mark-concurrent-migrations
  ef/validate-migration-cycles
  ef/refactor-runner
  ef/definitions-first
  ef/speeling
  ef/definitions-count
  ef/repair-testdata
  ef/speeling
  ef/extract-storetypes
  ef/backfill-migration-logs
  ef/update-memory-store
  ef/store-no-validate
  ef/update/desugar-revert
  ef/check-multiple-versions
  ef/unify-up-down
  ef/with-locked-schema-state
  ef/migration-graph
  ef/migrator-concurrent-index
  ef/remove-dead-code
)

pr_ids=(
  '29831'
  '30248'
  '30249'
  '30250'
  '30251'
  '30262'
  '30265'
  '30268'
  '30269'
  '30270'
  '30271'
  '30273'
  '30276'
  '30309'
  '30314'
  '30318'
  '30319'
  '30321'
  '30386'
  '30388'
  '30403'
  '30405'
  '30406'
  '30407'
  '30408'
  '30409'
  '30411'
  '30418'
  '30423'
  '30428'
  '30455'
  '30456'
  '30457'
  '30458'
  '30501'
  '30509'
  '30511'
  '30512'
  '30514'
  '30526'
  '30542'
  '30544'
  '30549'
  '30664'
  '30693'
  '30855'
)

echo "digraph git {"

i=0
for branch in "${branches[@]}"; do
  echo "  pr${i}[label=\"${branch}\"];"
  i=$((i + 1))
done

echo ""

i=0
main_commits=()
for branch in "${branches[@]}"; do
  prev_commit=$(git merge-base "$(git log "origin/${branch}" ^main --pretty='%H' | tac | head -n 1)" main)
  main_commits+=("${prev_commit}")

  for commit in $(git log "origin/${branch}" ^main --pretty='%H' | tac); do
    echo "  \"${commit}\" -> \"${prev_commit}\";"
    prev_commit=$commit
  done

  for commit in $(git log --all --grep "${pr_ids[i]}" --pretty='%H'); do
    if git merge-base --is-ancestor "${commit}" HEAD; then
      echo "  \"${commit}\" -> \"${prev_commit}\";"
      main_commits+=("${commit}")
    fi
  done

  echo "  \"pr${i}\" -> \"${prev_commit}\";"

  i=$((i + 1))
  echo ""
done

prev_commit="${main_commits[0]}"
for commit in $(git log "${main_commits[0]}"..HEAD --pretty='%H' | tac); do
  if [[ " ${main_commits[*]} " =~ ${commit} ]]; then
    echo "  \"${commit}\" -> \"${prev_commit}\";"
    prev_commit=$commit
  fi
done

echo "  \"main\" -> \"${prev_commit}\";"
echo "}"
