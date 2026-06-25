#!/usr/bin/env bash
# Check that no commit author/committer fields contain the word "paperclip" (case-insensitive).
# Used in CI to prevent accidental Paperclip references in git metadata.

set -euo pipefail

# Get commit range for the current PR (or latest commit on push)
if [[ "${GITHUB_EVENT_NAME:-}" == "pull_request" ]]; then
  RANGE="origin/${GITHUB_BASE_REF}..HEAD"
else
  RANGE="HEAD~1..HEAD"
fi

# Extract author and committer fields (name and email)
FIELDS=$(git log --format="an:%an ae:%ae cn:%cn ce:%ce" "$RANGE")

if echo "$FIELDS" | grep -i "paperclip"; then
  echo "ERROR: Git author/committer field contains 'paperclip'. Please update git config."
  echo "Matches:"
  echo "$FIELDS" | grep -i "paperclip"
  exit 1
fi

echo "OK: No Paperclip references found in author/committer fields."
