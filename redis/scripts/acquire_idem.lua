-- KEYS[1] = idem:{key}
-- ARGV[1] = ttlSec
-- ARGV[2] = reqHash ("" 가능)


local exists = redis.call("EXISTS", KEYS[1])
if exists == 1 then
    if ARGV[2] ~= '' then
        local saveHash = redis.call('HGET', KEYS[1], 'reqHash')
        if saveHash and saveHash ~= ARGV[2] then
            return {'MISMATCH', '', ''}
        end
    end
    local state = redis.call('HGET', KEYS[1], 'state')
    if state == 'done' then
        local status = redis.call('HGET', KEYS[1], 'status') or ''
        local body = redis.call('HGET', KEYS[1], 'body') or ''
        return {'DONE' , status, body}
    else
        return {'PROCESSING', '', ''}
    end
else
    redis.call('HSET', KEYS[1],
    'state', 'processing',
    'reqHash', ARGV[2],
)
redis.call('EXPIRE', KEYS[2], tonumber(ARGV[1]))
return {'LOCKED', '', ''}
end