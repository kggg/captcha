package captcha

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	MemoryStoreMode = NewMemoryStore(180)
)

type Storer interface {
	Save(id string, cap Captcha) error
	Get(id string) (string, error)
	Verify(id, code string) bool
	Remove(id string) error
}

type MemoryStore struct {
	sync.Mutex
	Value      map[string]Captcha
	Expiration time.Duration
	Total      int
}

func NewMemoryStore(expire time.Duration) Storer {
	m := new(MemoryStore)
	m.Value = make(map[string]Captcha)
	m.Expiration = expire
	m.Total = 0
	return m
}

func (m *MemoryStore) Save(id string, captcha Captcha) error {
	m.Lock()
	defer m.Unlock()
	m.Value[id] = captcha
	m.Total++
	return nil
}

func (m *MemoryStore) Get(id string) (string, error) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Value[id]; !ok {
		return "", fmt.Errorf("the %s not existing", id)
	}
	//计算保存时间是否已经超时
	now := time.Now()
	if now.Sub(m.Value[id].StartTime) > m.Expiration {
		delete(m.Value, id)
		return "", fmt.Errorf("the %s is expired", id)
	}
	answer := m.Value[id].Code
	return answer, nil
}

func (m *MemoryStore) Verify(id, code string) bool {
	if id == "" || code == "" {
		return false
	}
	v, err := m.Get(id)
	if err != nil {
		return false
	}
	return v != "" && v == code
}

func (m *MemoryStore) Remove(id string) error {
	delete(m.Value, id)
	return nil
}

type RedisStore struct {
	Client     *redis.Client
	Expiration time.Duration
}

var ctx = context.Background()

func NewRedisStore(host, password, port string, db int, expire time.Duration) Storer {
	addr := fmt.Sprintf("%s:%s", host, port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	r := new(RedisStore)
	r.Client = rdb
	r.Expiration = expire
	return r
}

func (r *RedisStore) Get(id string) (string, error) {
	res, err := r.Client.Get(ctx, id).Result()
	if err != nil {
		return "", fmt.Errorf("redis get %s error: %w", id, err)
	}
	return res, nil

}

func (r *RedisStore) Save(id string, cap Captcha) error {
	_, err := r.Client.SetNX(ctx, id, cap, 300).Result()
	if err != nil {
		return fmt.Errorf("redis set %s error: %w", id, err)
	}
	return nil
}

func (r *RedisStore) Verify(id, code string) bool {
	if id == "" || code == "" {
		return false
	}
	v, err := r.Get(id)
	if err != nil {
		fmt.Println("Err: ", err)
		return false
	}
	return v != "" && v == code
}

func (r *RedisStore) Remove(id string) error {
	_, err := r.Client.Del(ctx, id).Result()
	if err != nil {
		return fmt.Errorf("redis del %s error: %w", id, err)
	}
	return nil
}
