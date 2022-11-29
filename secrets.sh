#!/usr/bin/bash

REPO="rschmied/ghsecret"

# read the public github key for our repo
read -d' ' GH_KEY_ID GH_KEY <<< "$(gh api /repos/$REPO/actions/secrets/public-key | jq -r '.|.key_id, .key')"

# make them visible to the ghsecret tool
export GH_KEY GH_KEY_ID TUNNEL

if [ -z $1 ]; then
    echo "adding secrets"
    secret-tool lookup ghgpgkey value | \
        ghsecret | \
        gh api -XPUT /repos/$REPO/actions/secrets/GPG_PASSPHRASE --input -

    echo $GPG_PRIVATE_KEY | \
        ghsecret | \
        gh api -XPUT /repos/$REPO/actions/secrets/GPG_PRIVATE_KEY --input -
else
    echo "removing secrets"
    gh api -XDELETE /repos/$REPO/actions/secrets/GPG_PASSPHRASE
    gh api -XDELETE /repos/$REPO/actions/secrets/GPG_PRIVATE_KEY
fi
