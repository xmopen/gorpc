package netpoll

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	Tcp = 0
	Udp = 1
)

var (
	PoolMaxConnConfigError = fmt.Errorf("pool max conn config err")
)

var connIndex uint8 = 0x00 // increment id of conn.

// TODO: 第一版做的稍微简单一点即可.

// Pool 连接池,需要区分TCP和UDP.
type Pool interface {
	NewConn() (*Conn, error)
	CloseConn(conn *Conn) error

	Get() (*Conn, error)
	Put(conn *Conn)
	Remove(conn *Conn)

	Len() int16
	IdleLen() int16
	reaper()

	// Close 关闭连接池.
	Close() error
}

// PoolConfig 连接池配置.
type PoolConfig struct {
	PoolSize       int16
	IdleMinSize    int16
	PoolMaxSize    int16
	MaxIdleTimeout int64 // 最大空闲等待时间,单位秒.
}

// ConnPool 连接池具体实现.
type ConnPool struct {
	queue      *queue
	poolConfig *PoolConfig
	lock       sync.Locker
}

// Conn conn.
type Conn struct {
	conn         net.Conn
	lastUserTime int64 // 最近一次使用时间.
	state        int8
	cid          uint8 //一个字节即可.
}

func NewConnPool(cf *PoolConfig) (Pool, error) {
	pool := &ConnPool{
		queue:      newQueue(cf.PoolMaxSize),
		poolConfig: cf,
	}

	// 初始化conn.
	if cf.IdleMinSize > cf.PoolMaxSize {
		return nil, PoolMaxConnConfigError
	}
	if err := initPool(pool); err != nil {
		return nil, err
	}

	// 定期检查超过等待时间的conn,从而清除.
	go pool.reaper()
	return pool, nil
}

func initPool(pool *ConnPool) error {
	for i := 0; i < int(pool.poolConfig.PoolSize); i++ {
		conn, err := pool.NewConn()
		if err != nil {
			return err
		}
		pool.queue.Push(conn)
	}
	return nil
}

func (c *ConnPool) reaper() {

}

func (c *ConnPool) NewConn() (*Conn, error) {
	//TODO implement me
	// 1、时间.
	// 2、cid: connIndex++
	conn := &Conn{
		lastUserTime: time.Now().Unix(),
		cid:          connIndex + 1,
	}
	c.queue.Push(conn)
	panic("implement me")
}

func (c *ConnPool) CloseConn(conn *Conn) error {
	// 创建的时候应该把最近使用时间更新到.
	return nil
}

// Get get conn from pool.
func (c *ConnPool) Get() (*Conn, error) {
	conn := c.queue.Get()
	// 可能连接未到达最大连接数.
	if conn == nil {
		conn, err := c.NewConn()
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	conn.lastUserTime = time.Now().Unix()
	return conn, nil
}

// Put put a conn to queue.
func (c *ConnPool) Put(conn *Conn) {
	c.queue.Push(conn)
}

func (c *ConnPool) Remove(conn *Conn) {
	//c.queue.Delete()
}

func (c *ConnPool) Len() int16 {
	//TODO implement me
	panic("implement me")
}

func (c *ConnPool) IdleLen() int16 {
	//TODO implement me
	panic("implement me")
}

func (c *ConnPool) Close() error {
	//TODO implement me
	panic("implement me")
}
