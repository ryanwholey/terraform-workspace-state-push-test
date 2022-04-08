#!/bin/bash

src_id="ws-uzdMhTpBLpDqaKqW"
dst_id="ws-72cyhs4rXQHRAKAN"

src_base_url="https://app.terraform.io/api/v2"
dst_base_url="https://app.terraform.io/api/v2"

state_file_url=$(curl -fsSL \
  -H "Authorization: Bearer $TF_TOKEN" \
  -H "Content-Type: application/vnd.api+json" \
  "${src_base_url}/workspaces/${src_id}/current-state-version" | jq -r '.data.attributes["hosted-state-download-url"]')

state=$(curl $state_file_url)

body_file=$(mktemp)

cat <<EOF >> $body_file
{
  "data": {
    "type":"state-versions",
    "attributes": {
      "serial": 1,
      "md5": "$(echo -n $state | md5sum | awk '{print $1}')",
      "state": "$(echo $state)"
    }
  }
}
EOF

curl -fsSl "${dst_base_url}/workspaces/${dst_id}/actions/lock" \
  -X POST \
  -H "Authorization: Bearer ${TF_TOKEN}" \
  -H "Content-Type: application/vnd.api+json" \
  -d '{"reason": "Lock to push initial state"}'

curl "${dst_base_url}/workspaces/${dst_id}/state-versions" \
  -H "Authorization: Bearer $TF_TOKEN" \
  -H "Content-Type: application/vnd.api+json" \
  -d "@$body_file"

curl -fsSl "${dst_base_url}/workspaces/${dst_id}/actions/unlock" \
  -X POST \
  -H "Authorization: Bearer ${TF_TOKEN}" \
  -H "Content-Type: application/vnd.api+json"

# curl https://app.terraform.io/api/v2/workspaces/ws-72cyhs4rXQHRAKAN/state-versions \
#   -X POST \
#   -H "Authorization: Bearer $TF_TOKEN" \
#   -H "Content-Type: application/vnd.api+json" \
#   -d @- << EOF

# EOF
  
