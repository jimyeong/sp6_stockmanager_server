
-- KEYS[1] = idem:{key}
-- ARGV[1] = status
-- ARGV[2] = body
-- ARGV[3] = ttlSec

if redis.call("EXISTS", KEYS[1]) == 0 then
    return "MISSING"
end
redis.call("HSET", KEYS[1],
"state", "done", 
"status", ARGV[1],
"body", ARGV[2],
)
redis.call("EXPIRE", KEYS[1], tonumber(ARGV[3]))
return "OK"