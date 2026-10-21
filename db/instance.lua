#!/usr/bin/env tarantool
-- SPDX-License-Identifier: AGPL-3.0-or-later
-- Copyright (C) 2023 Dyne.org foundation <foundation@dyne.org>.
--
-- This program is free software: you can redistribute it and/or modify
-- it under the terms of the GNU Affero General Public License as
-- published by the Free Software Foundation, either version 3 of the
-- License, or (at your option) any later version.
--
-- This program is distributed in the hope that it will be useful,
-- but WITHOUT ANY WARRANTY; without even the implied warranty of
-- MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
-- GNU Affero General Public License for more details.
--
-- You should have received a copy of the GNU Affero General Public License
-- along with this program.  If not, see <https://www.gnu.org/licenses/>.

box.cfg {listen = 3500}
local function bootstrap()
    print("Start bootstrap")
    local tx_id = box.schema.sequence.create('tx_id',{start=0,min=0,step=1})
    local txs = box.schema.create_space('TXS', {engine = 'vinyl'})
    txs:create_index('primary', {sequence='tx_id'})
    txs:create_index('owner_token', { unique=false, parts = {
        {field = 2, type = 'string'},
        {field = 3, type = 'string'},
        {field = 4, type = 'unsigned'},
    }})
    txs:format({{name='PK', type='unsigned',is_nullable=false},
                {name='OWNER', type='string',is_nullable=false},
                {name='TOKEN', type='string',is_nullable=false},
                {name='TIMESTAMP', type='unsigned',is_nullable=false},
                {name='AMOUNT', type='string',is_nullable=false},
    })

    -- Keep things safe by default
    box.schema.user.create('wallet', { password = 'wallet' })
    box.schema.user.grant('wallet', 'replication')
    box.schema.user.grant('wallet', 'read,write,execute', 'space')
    box.schema.user.grant('wallet', 'read,write', 'sequence')
end
box.once('wallet-0', bootstrap)

-- Idempotency store for event-backed issuance: one row per paid event id, so
-- a point issuance mints at most once per ValueFlows/social event. New box.once
-- key so it provisions on existing instances too (wallet-0 never re-runs). The
-- universal space grant in bootstrap already covers this space.
local function bootstrap_payouts()
    print("Start bootstrap payouts")
    local payouts = box.schema.create_space('PAYOUTS', {engine = 'vinyl', if_not_exists = true})
    payouts:create_index('primary', {parts = {{field = 1, type = 'string'}}, if_not_exists = true})
    payouts:format({{name='EVENT_ID', type='string', is_nullable=false}})
end
box.once('wallet-1', bootstrap_payouts)

-- load my_app module and call start() function
-- with some app options controlled by sysadmins
local m = require('wallet').start()

