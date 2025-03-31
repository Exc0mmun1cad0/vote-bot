-- Create spaces --
box.schema.space.create('polls', { if_not_exists = true })
box.schema.space.create('options', { if_not_exists = true })
box.schema.space.create('votes', { if_not_exists = true })

-- Specify field names and types --
box.space.polls:format({
    {name = 'id', type = 'unsigned'},
    {name = 'poll_name', type = 'string'},
    {name = 'creator', type = 'string'},
    {name = 'channel', type = 'string'},
    {name = 'is_finished', type = 'boolean', default = false},
    {name = 'is_multi_vote', type = 'boolean'}
})

box.space.options:format({
    {name = 'id', type = 'unsigned'},
    {name = 'poll_id', type = 'unsigned'},
    {name = 'option_name', type = 'string'},
    {name = 'option_num', type = 'unsigned'}
})

box.space.votes:format({
    {name = 'id', type = 'unsigned'},
    {name = 'user', type = 'string'},
    {name = 'poll_id', type = 'unsigned'},
    {name = 'option_nums', type = 'array'}
})

-- Create sequences --
box.schema.sequence.create('poll_id', { if_not_exists = true })
box.schema.sequence.create('option_id', { if_not_exists = true })
box.schema.sequence.create('vote_id', { if_not_exists = true })

-- Primary
box.space.polls:create_index('primary', { parts = { 'id' }, sequence = 'poll_id', if_not_exists = true })
box.space.options:create_index('primary', { parts = { 'id' }, sequence = 'option_id', if_not_exists = true })
box.space.votes:create_index('primary', { parts = { 'id' }, sequence = 'vote_id', if_not_exists = true })

-- Secondary
box.space.options:create_index('option_poll_id', { unique = false, parts = { 'poll_id' }, if_not_exists = true })
box.space.votes:create_index('vote_poll_id', { unique = false, parts = { 'poll_id' }, if_not_exists = true })
box.space.votes:create_index('vote_user_poll_id', { unique = true, parts = {'user', 'poll_id'}, if_not_exists = true })

-- Add helper functions --
-- For options deletion
function delete_options(poll_id)
    for _, option in box.space.options.index.option_poll_id:pairs(poll_id) do
        box.space.options:delete{option.id}
    end
end

box.schema.func.create('delete_options', { if_not_exists = true })

-- For votes deletion
function delete_votes(poll_id)
    for _, option in box.space.votes.index.vote_poll_id:pairs(poll_id) do
        box.space.votes:delete{option.id}
    end
end

box.schema.func.create('delete_votes', { if_not_exists = true })

-- For votes creation
-- (in case vote exists it's updated)
function create_vote(user, poll_id, new_option_ids)
    local existing = box.space.votes.index.vote_user_poll_id:get{user, poll_id}

    if existing then
        box.space.votes:update(existing.id, {{'=', 4, new_option_ids}})
        return box.space.votes:get{existing.id}
    else 
        local new_id = box.sequence.vote_id:next()
        box.space.votes:insert{new_id, user, poll_id, new_option_ids}
        return box.space.votes:get{new_id}
    end
end

box.schema.func.create('create_vote', { if_not_exists = true })
