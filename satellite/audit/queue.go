// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package audit

import (
	"context"
	"sync"

	"github.com/zeebo/errs"
)

// ErrEmptyQueue is used to indicate that the queue is empty.
var ErrEmptyQueue = errs.Class("empty audit queue")

// Queue is a list of segments to audit, shared between the reservoir chore and audit workers.
// It is not safe for concurrent use.
type Queue struct {
	queue []Segment
}

// NewQueue creates a new audit queue.
func NewQueue(segments []Segment) *Queue {
	return &Queue{
		queue: segments,
	}
}

// Next gets the next item in the queue.
func (q *Queue) Next() (Segment, error) {
	if len(q.queue) == 0 {
		return Segment{}, ErrEmptyQueue.New("")
	}

	next := q.queue[0]
	q.queue = q.queue[1:]

	return next, nil
}

// Size returns the size of the queue.
func (q *Queue) Size() int {
	return len(q.queue)
}

// ErrPendingQueueInProgress means that a chore attempted to add a new pending queue when one was already being added.
var ErrPendingQueueInProgress = errs.Class("pending queue already in progress")

// Queues is a shared resource that keeps track of the next queue to be fetched
// and swaps with a new queue when ready.
type Queues struct {
	mu           sync.Mutex
	nextQueue    *Queue
	swapQueue    func()
	queueSwapped chan struct{}
}

// NewQueues creates a new Queues object.
func NewQueues() *Queues {
	queues := &Queues{
		nextQueue: NewQueue([]Segment{}),
	}
	return queues
}

// Fetch gets the active queue, clears it, and swaps a pending queue in as the new active queue if available.
func (queues *Queues) Fetch() *Queue {
	queues.mu.Lock()
	defer queues.mu.Unlock()

	if queues.nextQueue.Size() == 0 && queues.swapQueue != nil {
		queues.swapQueue()
	}
	active := queues.nextQueue

	if queues.swapQueue != nil {
		queues.swapQueue()
	} else {
		queues.nextQueue = NewQueue([]Segment{})
	}

	return active
}

// Push waits until the next queue has been fetched (if not empty), then swaps it with the provided pending queue.
// Push adds a pending queue to be swapped in when ready.
// If nextQueue is empty, it immediately replaces the queue. Otherwise it creates a swapQueue callback to be called when nextQueue is fetched.
// Only one call to Push is permitted at a time, otherwise it will return ErrPendingQueueInProgress.
func (queues *Queues) Push(pendingQueue []Segment) error {
	queues.mu.Lock()
	defer queues.mu.Unlock()

	// do not allow multiple concurrent calls to Push().
	// only one audit chore should exist.
	if queues.swapQueue != nil {
		return ErrPendingQueueInProgress.New("")
	}

	if queues.nextQueue.Size() == 0 {
		queues.nextQueue = NewQueue(pendingQueue)
		return nil
	}

	queues.queueSwapped = make(chan struct{})

	queues.swapQueue = func() {
		queues.nextQueue = NewQueue(pendingQueue)
		queues.swapQueue = nil
		close(queues.queueSwapped)
	}
	return nil
}

// WaitForSwap blocks until the swapQueue callback is called or context is canceled.
// If there is no pending swap, it returns immediately.
func (queues *Queues) WaitForSwap(ctx context.Context) error {
	queues.mu.Lock()
	if queues.swapQueue == nil {
		queues.mu.Unlock()
		return nil
	}
	queues.mu.Unlock()

	// wait for swapQueue to be called or for context canceled
	select {
	case <-queues.queueSwapped:
	case <-ctx.Done():
	}

	return ctx.Err()
}

// VerifyQueue controls manipulation of a database-based queue of segments to be
// verified; that is, segments chosen at random from all segments on the
// satellite, for which workers should perform audits. We will try to download a
// stripe of data across all pieces in the segment and ensure that all pieces
// conform to the same polynomial.
type VerifyQueue interface {
	Push(ctx context.Context, segments []Segment, maxBatchSize int) (err error)
	Next(ctx context.Context) (Segment, error)
}

// ReverifyQueue controls manipulation of a queue of pieces to be _re_verified;
// that is, a node timed out when we requested an audit of the piece, and now
// we need to follow up with that node until we get a proper answer to the
// audit. (Or until we try too many times, and disqualify the node.)
type ReverifyQueue interface {
	Insert(ctx context.Context, piece PieceLocator) (err error)
	GetNextJob(ctx context.Context) (job ReverificationJob, err error)
	Remove(ctx context.Context, piece PieceLocator) (wasDeleted bool, err error)
}

// ByStreamIDAndPosition allows sorting of a slice of segments by stream ID and position.
type ByStreamIDAndPosition []Segment

func (b ByStreamIDAndPosition) Len() int {
	return len(b)
}

func (b ByStreamIDAndPosition) Less(i, j int) bool {
	comparison := b[i].StreamID.Compare(b[j].StreamID)
	if comparison < 0 {
		return true
	}
	if comparison > 0 {
		return false
	}
	return b[i].Position.Less(b[j].Position)
}

func (b ByStreamIDAndPosition) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
