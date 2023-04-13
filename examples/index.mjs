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

const PIPPO_EDDSA = ...
const PIPPO_EDDSA_PK = ...
const PIPPO_ID = "IDDDPIPPO"

const url="http://localhost:8000"
//const url="https://gateway0.interfacer.dyne.org/wallet"

const signRequest = async (json, key, pk) => {
	const data = `{"gql": "${Buffer.from(json, 'utf8').toString('base64')}"}`
    const keys = `{"keyring": {"eddsa": "${key}"}}`
	const {result} = await zencode_exec(sign(), {data, keys});
	return {
		'did-sign': JSON.parse(result).eddsa_signature,
		'did-pk': pk,
	}
}

const addDiff = async () => {
    const request = {
	    token: "idea",
	    amount: '100',
	    owner: PIPPO_ID,
    }
    const requestJSON = JSON.stringify(request)
    const requestHeaders =  await signRequest(requestJSON, PIPPO_EDDSA, PIPPO_EDDSA_PK);
    const config = {
        headers: requestHeaders
    };

    const result = await axios.post(`${url}/token`, request, config);
    return result
}

const getToken = async () => {
    const request = {
	    token: "idea",
	    owner: PIPPO_ID,
    }

    const result = await axios.get(`${url}/token/${request.token}/${request.owner}`);
    return result
}

const getTxs = async () => {
    const request = {
	    token: "strength",
	    owner: PIPPO_ID,
    }

    const result = await axios.get(`${url}/token/${request.token}/${request.owner}/last/10`);
    return result
}
console.log(await addDiff())
console.log(await getToken())

