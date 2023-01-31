#!/usr/bin/env tarantool

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

-- load my_app module and call start() function
-- with some app options controlled by sysadmins
local m = require('wallet').start()

