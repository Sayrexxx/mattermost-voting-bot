box.cfg{
    listen = '3301',
    work_dir = '/var/lib/tarantool',
    memtx_dir = '/var/lib/tarantool',
    wal_dir = '/var/lib/tarantool',
    log_level = 5
}

box.schema.user.create('admin', {password = 'password'})
box.schema.user.grant('admin', 'read,write,execute', 'universe')

local votes = box.schema.create_space('votes', {
    if_not_exists = true,
    format = {
        {name = 'id', type = 'string'},
        {name = 'creator_id', type = 'string'},
        {name = 'channel_id', type = 'string'},
        {name = 'question', type = 'string'},
        {name = 'options', type = 'array'},
        {name = 'votes', type = 'map'},
        {name = 'created_at', type = 'string'},
        {name = 'closed_at', type = 'string', is_nullable = true}
    }
})

votes:create_index('primary', {
    type = 'hash',
    parts = {'id'},
    if_not_exists = true
})

votes:create_index('channel', {
    type = 'tree',
    parts = {'channel_id'},
    if_not_exists = true
})

votes:create_index('creator', {
    type = 'tree',
    parts = {'creator_id'},
    if_not_exists = true
})

box.schema.func.create('vote.add_vote', {
    if_not_exists = true,
    body = function(vote_id, user_id, option_idx)
        local vote = box.space.votes:get(vote_id)
        if not vote then
            return nil, 'Vote not found'
        end

        if vote.closed_at ~= nil then
            return nil, 'Vote is closed'
        end

        local votes = vote.votes or {}
        if votes[user_id] ~= nil then
            return nil, 'Already voted'
        end

        votes[user_id] = option_idx
        box.space.votes:update(vote_id, {{'=', 'votes', votes}})
        return true
    end
})