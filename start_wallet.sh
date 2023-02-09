#!/bin/sh

[ "$1" == "" ] && { >&2 echo "$0 token_name [other token names]"; exit 1;}

cd restroom

cp .env.example .env
# prepare .env file fro devops
path_to_zencode=`cat .env | grep "^ZENCODE_DIR=" | cut -d= -f 2`
devops_files_dir=`echo "$(dirname ${path_to_zencode})/devops_contracts"`
sed -i "s+ZENCODE_DIR=${path_to_zencode}+ZENCODE_DIR=${devops_files_dir}+g" .env

# create key and pk and start local restroom
yarn run init
pm2 start restroom.mjs --name=devops

# timestamp of creation
time=$(($(date +%s%N)/1000000))

for token in "$@"; do
    errtmp=`mktemp`
    restmp=`mktemp`

    # create tokens
    curl -X 'POST' -s -w "%{stderr}%{http_code}" \
    'http://localhost:3000/api/create_tokens' \
    -H 'accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{
    "data": {
        "amount": "100000000",
        "asset": {
            "timestamp": "'"${time}"'",
            "name": "'"${token}"'"
        }
    },
    "keys": {}
    }' > ${restmp} 2>${errtmp}
    # handle error
    if [ `cat ${errtmp}` != "200" ]; then
        printf "\nerror: "
        cat ${errtmp}
        cat ${restmp}
        echo
        rm -f ${errtmp} ${restmp}
        pm2 stop devops
        pm2 delete devops
        sed -i "s+ZENCODE_DIR=${path_to_zencode}+ZENCODE_DIR=${devops_files_dir}+g" .env
        exit 1
    fi
    # handle success
    txid=`cat ${restmp} | jq '.txid'`
    tmp=`mktemp` && jq --arg t "${token}" --arg id "${txid}" '.tokens[$t] = $id' ${path_to_zencode}/v1/_keys_setup.keys > ${tmp} && mv ${tmp} ${path_to_zencode}/v1/_keys_setup.keys
    rm -f ${errtmp} ${restmp}
done

# stop and delete restroom
pm2 stop devops
pm2 delete devops
# reset original .env
sed -i "s+ZENCODE_DIR=${devops_files_dir}+ZENCODE_DIR=${path_to_zencode}+g" .env
pm2 start restroom.mjs
