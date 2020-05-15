package redislock

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"
)

// defaultTimeout default lock release time.
const defaultTimeout = 5 * time.Minute
const unlockScript = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
const extendScript = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`

var pid = uint16(time.Now().UnixNano() & 65535)
var machineFlag uint16

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	machineFlag = hashNum(hostname)
	rand.Seed(time.Now().Unix())
}

func idGen() string {
	var b [16]byte
	binary.LittleEndian.PutUint16(b[:], pid)
	binary.LittleEndian.PutUint16(b[2:], machineFlag)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	binary.LittleEndian.PutUint32(b[12:], rand.Uint32())
	return base64.URLEncoding.EncodeToString(b[:])
}

func hashNum(str string) uint16 {
	tempv := int(str[0])
	for _, ruv := range str {
		tempv = 40503*tempv + int(ruv)
	}
	tempv &= 65535
	return uint16(tempv)
}

// IRedisClient Redis client interface.
type IRedisClient interface {
	RunExtendCmd(script, key, lockID string, expiration time.Duration) (err error, success bool)
	RunUnlockCmd(script, key, lockID string) (err error, success bool)
	SetNX(key, lockID string, expiration time.Duration) (err error, success bool)
}

// ILockFactory A lock factory interface.
type ILockFactory interface {
	GetLock(ctx context.Context, resourceID string) (ILock, error)
}

type lockFactory struct {
	redisClient IRedisClient
	lockPool    *sync.Pool
	idGen       func() string
}

// FactoryOptions custom params to replace default value
type FactoryOptions struct {
	// IDGenerator a function generate unique id for each lock, if not
	IDGenerator func() string
	// DefaultTimeout default lock release time, if not passed `DefaultTimeout` will be 5 minutes.
	DefaultTimeout time.Duration
}

// ILock Lock client interface.
type ILock interface {
	// Lock lock a resource, if failed will wait until success or reach the max times.
	Lock(ctx context.Context)
	// Realease release lock if lock id is matched.
	Release(ctx context.Context)
	// TryLock try to lock resource once.
	TryLock(ctx context.Context) (err error, success bool)
	// Extend extend lock time.
	Extend(ctx context.Context) (err error, success bool)
	// Get lock left seconds.
	TTL(ctx context.Context)
}

// NewLockFactory generate a lock factory.
func NewLockFactory(ctx context.Context, redisClient IRedisClient, options *FactoryOptions) (ILockFactory, error) {
	if redisClient == nil {
		panic(errors.New("redislock: no redis client"))
	}
	lockPool := &sync.Pool{
		New: func() interface{} {
			return &lock{
				redisClient: redisClient,
			}
		},
	}
	ret := lockFactory{
		redisClient: redisClient,
		lockPool:    lockPool,
	}

	if options.IDGenerator != nil {
		ret.idGen = options.IDGenerator
	}
	return &ret, nil
}

// GetLock Lock instance.
func (factory *lockFactory) GetLock(ctx context.Context, resourceID string) (ILock, error) {
	lock := factory.lockPool.Get().(*lock)
	lock.id = factory.idGen()
	return lock, nil
}

type lock struct {
	redisClient IRedisClient
	resourceID  string
	id          string
	expiration  time.Duration
}

func (l *lock) Lock(ctx context.Context) {

}

func (l *lock) Release(ctx context.Context) {

}

func (l *lock) TryLock(ctx context.Context) (err error, success bool) {
	return l.redisClient.SetNX(l.resourceID, l.id, l.expiration)
}

func (l *lock) Extend(ctx context.Context) (err error, success bool) {
	return nil, false
}

func (l *lock) TTL(ctx context.Context) {

}
