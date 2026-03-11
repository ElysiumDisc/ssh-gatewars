package store

import "time"

// ChatMsg mirrors the chat message row.
type ChatMsg struct {
	ID              int64
	Channel         string
	SenderFP        string
	SenderCallsign  string
	Kind            int
	Body            string
	CreatedAt       time.Time
}

// SaveChatMessage persists a chat message.
func (s *PlayerStore) SaveChatMessage(msg ChatMsg) error {
	_, err := s.db.Exec(`
		INSERT INTO chat_messages (channel, sender_fp, sender_callsign, kind, body, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, msg.Channel, msg.SenderFP, msg.SenderCallsign, msg.Kind, msg.Body, msg.CreatedAt)
	return err
}

// LoadBacklog loads the most recent messages for a channel.
func (s *PlayerStore) LoadBacklog(channel string, limit int) ([]ChatMsg, error) {
	rows, err := s.db.Query(`
		SELECT id, channel, sender_fp, sender_callsign, kind, body, created_at
		FROM chat_messages
		WHERE channel = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, channel, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []ChatMsg
	for rows.Next() {
		var m ChatMsg
		if err := rows.Scan(&m.ID, &m.Channel, &m.SenderFP, &m.SenderCallsign, &m.Kind, &m.Body, &m.CreatedAt); err != nil {
			continue
		}
		msgs = append(msgs, m)
	}

	// Reverse to chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// PruneChatMessages deletes messages older than the given time.
func (s *PlayerStore) PruneChatMessages(olderThan time.Time) error {
	_, err := s.db.Exec("DELETE FROM chat_messages WHERE created_at < ?", olderThan)
	return err
}
