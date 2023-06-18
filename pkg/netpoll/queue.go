package netpoll

import (
	"sync"
)

// conn queue 先进后出.
type queue struct {
	contains []*Conn
	size     int16
	index    int16 // 用于进行访问.
	head     int16 // 用于进行删除数据.
	tail     int16 // 用于进行添加数据.
	lock     sync.RWMutex

	indexMap map[uint8]int16
}

// newQueue return queue instance.
func newQueue(size int16) *queue {
	return &queue{
		contains: make([]*Conn, size),
		size:     0,
		index:    0,
		head:     0,
		tail:     0,
		indexMap: make(map[uint8]int16),
	}
}

// Push add conn.
func (q *queue) Push(cn *Conn) {
	if int(q.size) > len(q.contains) {
		return
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	q.contains[q.tail] = cn
	q.tail++
	q.size++
	q.indexMap[cn.cid] = q.tail - 1
}

// Get get conn.
func (q *queue) Get() *Conn {
	if q.head == q.tail {
		return nil
	}
	q.lock.Lock()
	conn := q.contains[q.index]
	q.index++
	q.index = (q.index + q.size) % q.size
	q.lock.Unlock()
	q.Delete(conn)
	return conn
}

// Delete remove conn from queue.
func (q *queue) Delete(conn *Conn) {
	if q.head == q.tail {
		return
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	q.contains[q.head] = nil
	for i := 1; i < int(q.size); i++ {
		q.contains[i-1] = q.contains[i]
	}
	q.contains[q.tail-1] = nil
	q.tail--
	q.size--
	q.index--
}
