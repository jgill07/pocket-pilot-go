#!/usr/bin/env bash
# scripts/worktree.sh <type> <desc>
#   e.g. scripts/worktree.sh feat add-transaction
#        -> branch feat/add-transaction, worktree .worktrees/feat-add-transaction
set -euo pipefail

types="feat fix docs refactor test chore build ci perf style"
type="${1:?usage: worktree.sh <type> <desc>   (types: $types)}"
desc="${2:?usage: worktree.sh <type> <desc>   (types: $types)}"

case " $types " in
  *" $type "*) ;;
  *) echo "invalid type '$type' — allowed: $types" >&2; exit 1 ;;
esac

branch="${type}/${desc}"
dir=".worktrees/${type}-${desc}"

git worktree add "$dir" -b "$branch"
[ -f .env ] && cp .env "$dir/.env"    # carry env into the worktree

echo "→ branch $branch"
echo "→ cd $dir"
