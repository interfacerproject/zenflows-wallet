import {
    promises as fsp
} from 'fs';
import fs from 'fs';
import path from 'path';
import dotenv from "dotenv";
import { zencode_exec } from "zenroom";

dotenv.config();

const ZENCODE_DIR = process.env.ZENCODE_DIR || "devops_contracts";
const FILES_DIR = process.env.FILES_DIR || "data";

const zen = async (zencode, keys, data) => {
    const params = {};
    if (keys !== undefined && keys !== null) {
        params.keys = typeof keys === 'string' ? keys : JSON.stringify(keys);
    }
    if (data !== undefined && data !== null) {
        params.data = typeof data === 'string' ? data : JSON.stringify(data);
    }
    try {
        return await zencode_exec(zencode, params);
    } catch (e) {
        console.log("Error from zencode_exec: ", e);
    }
}

// generate private key
let keyring = {}
if(!fs.existsSync(path.join(FILES_DIR, "keyring.json"))){
    const createKeyringScript = await fsp.readFile(path.join(ZENCODE_DIR, "create_keyring.zen"), 'utf8');
    const keys = await zen(createKeyringScript, null, null);
    if (!keys) {
	console.error("Error in generate private keys");
	process.exit(-1);
    }

    Object.assign(keyring, JSON.parse(keys.result));
    await fsp.writeFile(
	path.join(FILES_DIR, "keyring.json"),
	JSON.stringify(keyring), {mode: 0o600});
}
else {
    keyring = await fsp.readFile(path.join(FILES_DIR, "keyring.json"), 'utf8');
}

// generate public keys
const generatePublicKeysScript = await fsp.readFile(path.join(ZENCODE_DIR, "create_public_key.zen"), 'utf8');
let publicKeys = {};
const pubKeys = await zen(generatePublicKeysScript, keyring, null);
if (!pubKeys) {
    console.error("Error in generate public keys");
    process.exit(-1);
}

try {
    await fsp.unlink(path.join(FILES_DIR, "public_keys.json"));
} catch(e) {}
Object.assign(publicKeys, JSON.parse(pubKeys.result));
await fsp.writeFile(
    path.join(FILES_DIR, "public_keys.json"),
    JSON.stringify(publicKeys), {mode: 0o600});
