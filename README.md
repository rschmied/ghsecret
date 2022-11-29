[![CodeQL](https://github.com/rschmied/ghsecret/actions/workflows/codeql-analysis.yml/badge.svg?branch=main)](https://github.com/rschmied/ghsecret/actions/workflows/codeql-analysis.yml)

# README.md

This tool creates [libsodium](https://github.com/jedisct1/libsodium) compatible [encrypted secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets#creating-encrypted-secrets-for-an-environment) for use with the Github
CLI.  It takes the repo public key information that can be obtained via

    /repos/{owner}/{repo}/actions/secrets/public-key

which looks like

    {
        "key_id": "012345678912345678",
        "key": "2Sg8iYjAxxmI2LvUXpJjkYrMxURPc8r+dB7TJyvv1234"
    }

and a secret value provided via environment variable or standard input and creates the proper JSON
to be used with the Github CLI tool `gh`.

The required data looks like this:

    {
        "key_id":"012345678912345678"
        "encrypted_value":"base64encoded_secret",
    }

The tool prints the JSON object that can then be fed into `gh` like this:

    ghsecret ENVVARNAME | gh api -XPUT /some/endpoint --input -

`ENVVARNAME` is the name of the environment variable that holds the secret string.  In addition, the result from the "public-key" API call must to be provided as environment variables `GH_KEY` and `GH_KEY_ID`, respectively.

If `ENVVARNAME` is omitted, a value provided via stdin is used.

The names of these variables can be changed via command line arguments, if needed.

Here's a silly example:

```bash
$ GH_KEY="$(echo 'qwe' | base64)" GH_KEY_ID="123" ghsecret HOME
{"key_id":"123","encrypted_value":"bnwu9dXlXcFGYatcXsdpHR0MiiAE3115Mz6wkDrdNACQZSo+1JgPHrhaJCEEnbVpGF5YJMa3tJGGyeb2vqY="}
$
```

Here's another example using the Gnome key-chain `secret-tool`. This looks up the password in the login key-chain of Gnome where the associated key/value pair is "dc01admin" and "secret":

```bash
$ export GH_KEY="$(echo 'qwe' | base64)"
$ export GH_KEY_ID="123"
$ secret-tool lookup dc01admin secret | ghsecret
{"key_id":"123","encrypted_value":"ymkvNcCZ3Bykk6OtOa3csCJNmpdt2J9JA/iTYIASHn9L35UyuN+bzQuE6XhYHQWH3vNMy+FrDg=="}
$
```

Also see the [`secrets.sh`](secrets.sh) script in this repo for a more elaborate example.

## Prerequisites

The tool does one thing and one thing only: provide the input data that is understood by the Github secrets API. To actually consume this data, other tools are needed:

- The `gh` tool available from [here](https://cli.github.com/).
- The `jq` tool available as part of your distribution or from [here](https://stedolan.github.io/jq/).
- The Gnome `secret-tool` to retrieve secrets from the Gnome key-chain.

For my particular workflow, I also combine this with `tmux` and `ngrok` to allow access to my local installation of a service for Github actions.  

## Example

Here's some bash snippet that illustrates the actual / embedded use with a [PyPI publish workflow](https://github.com/marketplace/actions/pypi-publish):

```bash
REPO="rschmied/ghsecret"

# read the public github key for our repo
read -d' ' GH_KEY_ID GH_KEY <<< "$(gh api /repos/$REPO/actions/secrets/public-key | jq -r '.|.key_id, .key')"

# make them visible to the ghsecret tool
export GH_KEY GH_KEY_ID

# lookup the pypi token in the Gnome key-chain and set the it in the
# repository so that it can be consumed by a PyPI publish workflow.
secret-tool lookup token pypi | \
 ghsecret | \
 gh api -XPUT /repos/$REPO/actions/secrets/PYPI_API_TOKEN --input -
```

**Note:** Using `gh api -XDELETE /repos/$REPO/actions/secrets/PYPI_API_TOKEN` allows you to remove the token from Github again.
