#!/bin/bash

curl --location 'https://app.terraform.io/api/v2/organizations/servicevic/registry-providers' \
--header 'Content-Type: application/vnd.api+json' \
--header "Authorization: Bearer $TERRAFORM_CLOUD_TOKEN" \
--data '{
    "data": {
        "type": "registry-providers",
        "attributes": {
            "name": "jira",
            "namespace": "servicevic",
            "registry-name": "private"
        }
    }
}'
