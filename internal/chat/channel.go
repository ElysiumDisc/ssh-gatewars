package chat

// Channel represents a chat channel.
type Channel struct {
	Key     string
	Type    ChannelType
	Members map[string]bool // fingerprint set
	Backlog *RingBuffer
}

// NewChannel creates a channel with a backlog of the given size.
func NewChannel(key string, ctype ChannelType, backlogSize int) *Channel {
	return &Channel{
		Key:     key,
		Type:    ctype,
		Members: make(map[string]bool),
		Backlog: NewRingBuffer(backlogSize),
	}
}

// RingBuffer is a fixed-size circular buffer of chat messages.
type RingBuffer struct {
	buf  []ChatMessage
	cap  int
	head int // next write position
	size int
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &RingBuffer{
		buf: make([]ChatMessage, capacity),
		cap: capacity,
	}
}

// Push adds a message to the buffer, overwriting the oldest if full.
func (rb *RingBuffer) Push(msg ChatMessage) {
	rb.buf[rb.head] = msg
	rb.head = (rb.head + 1) % rb.cap
	if rb.size < rb.cap {
		rb.size++
	}
}

// All returns all messages in chronological order.
func (rb *RingBuffer) All() []ChatMessage {
	if rb.size == 0 {
		return nil
	}
	result := make([]ChatMessage, rb.size)
	start := (rb.head - rb.size + rb.cap) % rb.cap
	for i := 0; i < rb.size; i++ {
		result[i] = rb.buf[(start+i)%rb.cap]
	}
	return result
}

// Len returns the number of messages in the buffer.
func (rb *RingBuffer) Len() int {
	return rb.size
}
