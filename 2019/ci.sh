#!/bin/bash

set -e

# Check config file is valid
promtool check config config.yml

# Check alert and rule files are valid
promtool check rules alert*.yml rule*.yml

# Run all prometheus unit tests
promtool test rules test*.yml

# Ensure every alert has an assignee
yq -r -e "$(cat <<EOF

  .groups[].rules[] 
  | select(.labels.assignee == null)
  | .alert + " is missing an assignee"

EOF
)" alert*.yml && exit 1
