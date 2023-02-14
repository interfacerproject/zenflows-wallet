#!/usr/bin/env node

// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2023 Dyne.org foundation <foundation@dyne.org>.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

import sign from "./sign_graphql.mjs"
import { zencode_exec } from 'zenroom';
import axios from 'axios';

const PIPPO_EDDSA = "EtJtSqAG9mVHfKrKduS6aeyAE6okGXrfMW8fEQ6eqenh"
const PIPPO_ID = "062TE0H7591KJCVT3DDEMDBF0R"
const PLUTO_EDDSA = "2n4TEhoQ8ZwedJoUuJNbxv5W1cr5wHFYPcQmkk1EWj4t"
const PLUTO_ID = "062TE0YPJD392CS1DPV9XWMDXC"
const PAPERINO_EDDSA = "H7sbugVBZbmRX6M75WpzCi5vVVtaxvfLhovDijRAnZj"
const PAPERINO_ID = "062TE18QJSQJ1PY6G1M7783148"

const url="http://localhost:8000"
//const url="https://gateway0.interfacer.dyne.org/wallet"

const signRequest = async (json, key, id) => {
	const data = `{"gql": "${Buffer.from(json, 'utf8').toString('base64')}"}`
    const keys = `{"keyring": {"eddsa": "${key}"}}`
	const {result} = await zencode_exec(sign(), {data, keys});
	return {
		'zenflows-sign': JSON.parse(result).eddsa_signature,
		'zenflows-id': id,
	}
}

const sendMessage = async () => {
    const request = {
	    token: "idea",
	    amount: 100,
	    owner: "062TE0H7591KJCVT3DDEMDBF0R",
    }
    const requestJSON = JSON.stringify(request)
    const requestHeaders =  await signRequest(requestJSON, PIPPO_EDDSA, PIPPO_ID);
    const config = {
        headers: requestHeaders
    };

    const result = await axios.post(`${url}/token`, request, config);
    return result
}

const getToken = async () => {
    const request = {
	    token: "idea",
	    owner: "062TE0H7591KJCVT3DDEMDBF0R",
    }

    const result = await axios.get(`${url}/token/${request.token}/${request.owner}?until=1675694839000`);
    return result
}

const getTxs = async () => {
    const request = {
	    token: "idea",
	    owner: "062TE0H7591KJCVT3DDEMDBF0R",
    }

    const result = await axios.get(`${url}/token/${request.token}/${request.owner}/last/10`);
    return result
}
console.log(await sendMessage())
console.log(await getToken())
console.log((await getTxs()).data)

