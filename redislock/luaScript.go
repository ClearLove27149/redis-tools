package redislock

// lua script, atomic
const (
	compareAndDeleteScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`

	// cmd key, time, uuid
	repeatLockScript = `
	if (redis.call('exists', KEYS[1]) == 0) then
		redis.call('hset', KEYS[1], ARGV[2], 1)
		redis.call('pexpire', KEYS[1], ARGV[1])
		return 1
	end
	if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
		redis.call('hincrby', KEYS[1], ARGV[2], 1)
		redis.call('pexpire', KEYS[1], ARGV[1])
		return 2
	end
	return redis.call('pttl', KEYS[1])
	`

	// cmd key, time, uuid
	repeatUnLockScript = `
	if (redis.call('exists', KEYS[1]) == 0) then
		return 0
	end
	if (redis.call('hexists', KEYS[1], ARGV[2]) == 0) then
		return 0
	end
	local counter = redis.call('hincrby', KEYS[1], ARGV[2], -1)
	if (counter > 0) then
		redis.call('pexpire', KEYS[1], ARGV[1])
		return 1
	else
		redis.call('del', KEYS[1])
		return 2
	end
	`

	repeatLockSubScript = `
	if (redis.call('exists', KEYS[1]) == 0) then
		redis.call('hset', KEYS[1], ARGV[2], 1)
		redis.call('pexpire', KEYS[1], ARGV[1])
		return 1
	end
	if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
		redis.call('hincrby', KEYS[1], ARGV[2], 1)
		redis.call('pexpire', KEYS[1], ARGV[1])
		return 2
	end
	return redis.call('pttl', KEYS[1])
	`
	// 释放锁后，给topic发消息"ok"，通知其他客户端加锁
	repeatUnLockPubScript = `
	if (redis.call('exists', KEYS[1]) == 0) then
		return 0
	end
	if (redis.call('hexists', KEYS[1], ARGV[3]) == 0) then
		return 0
	end
	local counter = redis.call('hincrby', KEYS[1], ARGV[3], -1)
	if (counter > 0) then
		redis.call('pexpire', KEYS[1], ARGV[2])
		return 1
	else
		redis.call('del', KEYS[1])
		redis.call('publish', KEYS[2], "ok")
		return 2
	end
	`

	success = "OK"
	topic   = "pubsub"
)
