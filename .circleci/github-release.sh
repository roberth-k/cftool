#!/usr/bin/env bash
set -eu

is_prerelease="false"

if [[ "${CIRCLE_TAG}" =~ -[ab][0-9]+$ ]]; then
    echo "this is a pre-release"
    is_prerelease="true"
fi

release=$( \
    curl "https://api.github.com/repos/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}/releases" \
        -XPOST \
        -H "authorization: token ${GH_ACCESS_TOKEN}" \
        -H "content-type: application\json" \
        -d "{
            \"tag_name\":\"${CIRCLE_TAG}\",
            \"name\": \"${CIRCLE_TAG}\",
            \"draft\": true,
            \"prerelease\": ${is_prerelease}
        }" \
)

echo "${release}"

# Construct the upload url manually to avoid parsing parameter hints.
release_id=$(echo $release | jq -r ".id")
upload_url="https://uploads.github.com/repos/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}/releases/${release_id}/assets"
echo "upload_url=${upload_url}"

upload_dir=$(mktemp -d)

for dir in $(find $(pwd)/.build -mindepth 1 -type d); do
    cd $dir
    label=$(basename $dir)
    zip $upload_dir/$label.zip ./*
done

cd $upload_dir

for file in $(find . -name '*.zip'); do
    name=$(basename $file)
    curl "${upload_url}?name=${name}" \
        -XPOST \
        -H "authorization: token ${GH_ACCESS_TOKEN}" \
        -H "content-type: application/zip" \
        --data-binary @$file
done
