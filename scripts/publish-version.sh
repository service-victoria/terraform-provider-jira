#!/bin/bash

# Assumes dist folder exists and is populated with build artifacts
# Requires GPG_FINGERPRINT and TERRAFORM_CLOUD_TOKEN
VERSION=$(jq -r '.version' dist/metadata.json)

# Set GPG key used for signing
echo "Uploading GPG key"
gpg_response=$(curl --silent --location 'https://app.terraform.io/api/registry/private/v2/gpg-keys' \
--header 'Content-Type: application/vnd.api+json' \
--header "Authorization: Bearer $TERRAFORM_CLOUD_TOKEN" \
--data "{
    \"data\": {
        \"type\": \"gpg-keys\",
        \"attributes\": {
            \"namespace\": \"servicevic\",
            \"ascii-armor\": $(gpg --armor --export $GPG_FINGERPRINT | jq -R -s '.')
        }
    }
}")

# Create new provider version
echo "Creating new provider version"
version_response=$(curl --silent --location "https://app.terraform.io/api/v2/organizations/servicevic/registry-providers/private/servicevic/jira/versions" \
--header "Content-Type: application/vnd.api+json" \
--header "Authorization: Bearer $TERRAFORM_CLOUD_TOKEN" \
--data "{
    \"data\": {
        \"type\": \"registry-provider-versions\",
        \"attributes\": {
            \"version\": \"$VERSION\",
            \"key-id\": \"$GPG_FINGERPRINT\",
            \"protocols\": [\"6.0\"]
        }
    }
}")

echo "Uploading shasum and sig"
shasums_upload_url=$(jq -r '.data.links."shasums-upload"' <<< $version_response)
shasums_sig_upload_url=$(jq -r '.data.links."shasums-sig-upload"' <<< $version_response)

curl -s -T dist/*_SHA256SUMS "$shasums_upload_url"
curl -s -T dist/*_SHA256SUMS.sig "$shasums_sig_upload_url"

echo "Uploading artifacts"
jq -c '.[] | select(.type=="Archive")' dist/artifacts.json | while read i; do
  echo "Uploading $(jq -r '.name' <<< $i)"
  platform_response=$(curl --silent --location "https://app.terraform.io/api/v2/organizations/servicevic/registry-providers/private/servicevic/jira/versions/$VERSION/platforms" \
  --header "Content-Type: application/vnd.api+json" \
  --header "Authorization: Bearer $TERRAFORM_CLOUD_TOKEN" \
  --data "{
  \"data\": {
    \"type\": \"registry-provider-version-platforms\",
    \"attributes\": {
      \"os\": \"$(jq -r '.goos' <<< $i)\",
      \"arch\": \"$(jq -r '.goarch' <<< $i)\",
      \"shasum\": \"$(jq -r '.extra.Checksum' <<< $i)\",
      \"filename\": \"$(jq -r '.name' <<< $i)\"
    }
  }
}")

  provider_binary_upload_url=$(jq -r '.data.links."provider-binary-upload"' <<< $platform_response);
  curl -s -T $(jq -r '.path' <<< $i) "$provider_binary_upload_url";
done
